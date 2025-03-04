#include "MainServer.h"
#include "UserClientManager.h"

#include "Logger.h"
#include "ClientHeader.h"
#include "ThreadWorker.h"
#include "JobPool.h"
#include "Job.h"
#include "JobPriorities.h"

#include "fmt/format.h"
#include "fmt/xchar.h"

#include "boost/json.hpp"
#include "boost/json/parse.hpp"

#include <vector>
#include <filesystem>

using namespace Utilities;

MainServer::MainServer(std::shared_ptr<Configurations> configurations)
	: server_(nullptr)
	, thread_pool_(nullptr)
	, configurations_(configurations)
	, register_key_("MainServer")
	, redis_client_(nullptr)
{
	server_ = std::make_shared<NetworkServer>(configurations->client_title(), configurations->high_priority_count(), configurations->normal_priority_count(), configurations->low_priority_count());
	
	server_->register_key(register_key_);
	server_->received_connection_callback(std::bind(&MainServer::received_connection, this, std::placeholders::_1, std::placeholders::_2, std::placeholders::_3));
	server_->received_message_callback(std::bind(&MainServer::received_message, this, std::placeholders::_1, std::placeholders::_2, std::placeholders::_3));

	messages_.insert({ "request_client_status_update", std::bind(&MainServer::request_client_status_update, this, std::placeholders::_1, std::placeholders::_2, std::placeholders::_3) });
}

MainServer::~MainServer(void)
{
	stop();

	destroy_thread_pool();
}

auto MainServer::start() -> std::tuple<bool, std::optional<std::string>>
{
	auto [result, error_message] = create_thread_pool();
	if (!result)
	{
		Logger::handle().write(LogTypes::Error, fmt::format("Failed to create thread pool: {}", error_message.value()));
		return { false, fmt::format("Failed to create thread pool: {}", error_message.value()) };
	}

	if (configurations_->use_redis())
	{
		TLSOptions tls_options;
		tls_options.use_tls(configurations_->use_redis_tls());
		tls_options.ca_cert(configurations_->ca_cert());
		tls_options.client_cert(configurations_->client_cert());
		tls_options.client_key(configurations_->client_key());

		redis_client_ = std::make_shared<RedisClient>(configurations_->redis_host(), configurations_->redis_port(), tls_options, configurations_->redis_db_global_message_index());
		
		auto [connected, connect_error] = redis_client_->connect();
		if (!connected)
		{
			destroy_thread_pool();
			redis_client_.reset();

			Logger::handle().write(LogTypes::Error, fmt::format("Failed to connect redis: {}", connect_error.value()));
			return { false, fmt::format("Failed to connect redis: {}", connect_error.value()) };
		}

		redis_client_->set(global_message_key_, "");
	}

	std::tie(result, error_message) = server_->start(configurations_->server_port(), configurations_->buffer_size());
	if (!result)
	{
		Logger::handle().write(LogTypes::Error, fmt::format("Failed to start server: {}", error_message.value()));
		return { false, fmt::format("Failed to start server: {}", error_message.value()) };
	}

	auto [db_result, db_error] = thread_pool_->push(
		std::make_shared<Job>(JobPriorities::Low, std::bind(&MainServer::db_periodic_update_job, this), "db_periodic_update_job"));

	if (!db_result)
	{
		Logger::handle().write(LogTypes::Error, fmt::format("Failed to start db periodic update job: {}", db_error.value()));
		return { false, fmt::format("Failed to start db periodic update job: {}", db_error.value()) };
	}

	auto [consume_result, consume_error] = thread_pool_->push(
			std::make_shared<Job>(JobPriorities::High, std::bind(&MainServer::check_global_message, this), "check_global_message"));

	if (!consume_result)
	{
		Logger::handle().write(LogTypes::Error, fmt::format("Failed to start consume global message job: {}", consume_error.value()));
		return { false, fmt::format("Failed to start consume global message job: {}", consume_error.value()) };
	}

#ifdef WIN32
	system(fmt::format("MainServerConsumer --client_title {}", "MainServerConsumer").c_str());
#else
	system(fmt::format("./MainServerConsumer --client_title {}", "MainServerConsumer").c_str());
#endif

	return { true, std::nullopt };
}

auto MainServer::stop() -> void
{
	if (server_ == nullptr)
	{
		Logger::handle().write(LogTypes::Error, "server is null");
		return;
	}

	server_->stop();
	server_.reset();
}

auto MainServer::wait_stop() -> std::tuple<bool, std::optional<std::string>>
{
	if (server_ == nullptr)
	{
		Logger::handle().write(LogTypes::Error, "server is null");
		return { false, "server is null" };
	}

	return server_->wait_stop();
}

auto MainServer::create_thread_pool() -> std::tuple<bool, std::optional<std::string>>
{
	destroy_thread_pool();

	try
	{	
		thread_pool_ = std::make_shared<ThreadPool>();
	}
	catch(const std::bad_alloc& e)
	{
		return { false, fmt::format("Memory allocation failed to ThreadPool: {}", e.what()) };
	}
	
	for (auto i = 0; i < configurations_->high_priority_count(); i++)
	{
		std::shared_ptr<ThreadWorker> worker;
		try
		{
			worker = std::make_shared<ThreadWorker>(std::vector<JobPriorities>{ JobPriorities::High });
		}
		catch(const std::bad_alloc& e)
		{
			return { false, fmt::format("Memory allocation failed to ThreadWorker: {}", e.what()) };
		}

		thread_pool_->push(worker);
	}

	for (auto i = 0; i < configurations_->normal_priority_count(); i++)
	{
		std::shared_ptr<ThreadWorker> worker;
		try
		{
			worker = std::make_shared<ThreadWorker>(std::vector<JobPriorities>{ JobPriorities::Normal, JobPriorities::High });
		}
		catch(const std::bad_alloc& e)
		{
			return { false, fmt::format("Memory allocation failed to ThreadWorker: {}", e.what()) };
		}

		thread_pool_->push(worker);
	}

	for (auto i = 0; i < configurations_->low_priority_count(); i++)
	{
		std::shared_ptr<ThreadWorker> worker;
		try
		{
			worker = std::make_shared<ThreadWorker>(std::vector<JobPriorities>{ JobPriorities::Low });
		}
		catch(const std::bad_alloc& e)
		{
			return { false, fmt::format("Memory allocation failed to ThreadWorker: {}", e.what()) };
		}

		thread_pool_->push(worker);
	}

	auto [result, message] = thread_pool_->start();
	if (!result)
	{
		Logger::handle().write(LogTypes::Error, fmt::format("Failed to start thread pool: {}", message.value()));
		return { false, message.value() };
	}

	return { true, std::nullopt };
}

auto MainServer::destroy_thread_pool() -> void
{
	if (thread_pool_ == nullptr)
	{
		return;
	}

	thread_pool_->stop();
	thread_pool_.reset();
}

auto MainServer::received_connection(const std::string& id, const std::string& sub_id, const bool& condition) -> std::tuple<bool, std::optional<std::string>>
{
	if (server_ == nullptr)
	{
		Logger::handle().write(LogTypes::Error, "server is null");
		return { false, "server is null" };
	}

	if (thread_pool_ == nullptr)
	{
		Logger::handle().write(LogTypes::Error, "thread_pool is null");
		return { false, "thread_pool is null" };
	}

	if (condition)
	{
		Logger::handle().write(LogTypes::Information, fmt::format("Received connection[{}, {}]: connected", id, sub_id));
		
		UserClientManager::handle().add(id, sub_id);
		return { true, std::nullopt };
	}

	Logger::handle().write(LogTypes::Information, fmt::format("Received connection[{}, {}]: disconnected", id, sub_id));
	
	UserClientManager::handle().remove(id, sub_id);
	return { true, std::nullopt };
}

auto MainServer::received_message(const std::string& id, const std::string& sub_id, const std::string& message) -> std::tuple<bool, std::optional<std::string>>
{
	if (server_ == nullptr)
	{
		Logger::handle().write(LogTypes::Error, "server is null");
		return { false, "server is null" };
	}

	if (thread_pool_ == nullptr)
	{
		Logger::handle().write(LogTypes::Error, "thread_pool is null");
		return { false, "thread_pool is null" };
	}

	if (message.empty())
	{
		Logger::handle().write(LogTypes::Error, "message is empty");
		return { false, "message is empty" };
	}

	Logger::handle().write(LogTypes::Information, fmt::format("Received message[{}, {}]: {}", id, sub_id, message));

	return thread_pool_->push(
		std::dynamic_pointer_cast<Job>(
			std::make_shared<ClientMessageParsing>(
				id, sub_id, message, std::bind(&MainServer::parsing_message, this, std::placeholders::_1, std::placeholders::_2, std::placeholders::_3, std::placeholders::_4)
			)
		)
	);
}

auto MainServer::send_message(const std::string& message, const std::string& id, const std::string& sub_id) -> std::tuple<bool, std::optional<std::string>>
{
	if (server_ == nullptr)
	{
		Logger::handle().write(LogTypes::Error, "server is null");
		return { false, "server is null" };
	}

	Logger::handle().write(LogTypes::Information, fmt::format("Send message[{}, {}]: {}", id, sub_id, message));

	return server_->send_message(message, id, sub_id);
}

auto MainServer::parsing_message(const std::string& id, const std::string& sub_id, const std::string& command, const std::string& message) -> std::tuple<bool, std::optional<std::string>>
{
	if (command.empty())
	{
		Logger::handle().write(LogTypes::Error, "command is empty");
		return { false, "command is empty" };
	}

	if (message.empty())
	{
		Logger::handle().write(LogTypes::Error, "message is empty");
		return { false, "message is empty" };
	}

	if (thread_pool_ == nullptr)
	{
		Logger::handle().write(LogTypes::Error, "thread_pool is null");
		return { false, "thread_pool is null" };
	}

	auto iter = messages_.find(command);
	if (iter == messages_.end())
	{
		Logger::handle().write(LogTypes::Error, fmt::format("command is not found: {}", command));
		return { false, "command is not found" };
	}

	return thread_pool_->push(
		std::dynamic_pointer_cast<Job>(
			std::make_shared<ClientMessageExecute>(
				id, sub_id, message, iter->second
			)
		)
	);
}

auto MainServer::db_periodic_update_job() -> std::tuple<bool, std::optional<std::string>>
{
	if (thread_pool_ == nullptr)
	{
		Logger::handle().write(LogTypes::Error, "thread_pool is null");
		return { false, "thread_pool is null" };
	}

	auto clinets = UserClientManager::handle().clinets();
	boost::json::array user_list;
	for (const auto& [user_id, user_status] : clinets)
	{
		auto [id, sub_id] = user_id;
		auto [status, _] = user_status;
		boost::json::object user_object =
		{
			{ "id", id },
			{ "sub_id", sub_id },
			{ "status", status }
		};
		user_list.push_back(user_object);
	};

#ifdef WIN32
	system(fmt::format("db_cli --update --json_script {}", boost::json::serialize(user_list)).c_str());
#else
	system(fmt::foramt("./db_cli --update --json_script {}", boost::json::serialize(user_list)).c_str());
#endif

	auto job_pool = thread_pool_->job_pool();
	if (job_pool == nullptr || job_pool->lock())
	{
		Logger::handle().write(LogTypes::Error, "job_pool is null");
		return { false, "job_pool is null" };
	}

#ifdef _DEBUG
	std::this_thread::sleep_for(std::chrono::milliseconds(100));
#endif

	job_pool->push(
		std::make_shared<Job>(JobPriorities::Low, std::bind(&MainServer::db_periodic_update_job, this), "db_periodic_update_job"));

	return { true, std::nullopt };
}

auto MainServer::consume_message_queue() -> std::tuple<bool, std::optional<std::string>>
{
	if (redis_client_ == nullptr)
	{
		Logger::handle().write(LogTypes::Error, "redis_client is null");
		return { false, "redis_client is null" };
	}

	// static redis key polling
	auto [result, error_message] = redis_client_->get(global_message_key_);
	if (result.empty())
	{
		if (error_message.has_value())
		{
			Logger::handle().write(LogTypes::Error, fmt::format("Failed to get global message: {}", error_message.value()));
			return { false, fmt::format("Failed to get global message: {}", error_message.value()) };
		}

		Logger::handle().write(LogTypes::Sequence, "No global message");

		return { true, std::nullopt };
	}

	auto message_value = boost::json::parse(result);
	if (!message_value.is_object())
	{
		Logger::handle().write(LogTypes::Error, fmt::format("Failed to parse message: {}", result));
		return { false, "Failed to parse message" };
	}

	auto received_message = message_value.as_object();
	if (!received_message.contains("id") || !received_message.at("id").is_string())
	{
		Logger::handle().write(LogTypes::Error, fmt::format("Failed to parse message: {}", result));
		return { false, "Failed to parse message" };
	}

	if (!received_message.contains("sub_id") || !received_message.at("sub_id").is_string())
	{
		Logger::handle().write(LogTypes::Error, fmt::format("Failed to parse message: {}", result));
		return { false, "Failed to parse message" };
	}

	if (!received_message.contains("message") || !received_message.at("message").is_string())
	{
		Logger::handle().write(LogTypes::Error, fmt::format("Failed to parse message: {}", result));
		return { false, "Failed to parse message" };
	}

	boost::json::object message_object =
	{
		{ "id", received_message.at("id").as_string().data() },
		{ "sub_id", received_message.at("sub_id").as_string().data() },
		
		{ "data", received_message.at("message").as_string().data() }
	};

	boost::json::object broadcast_message = 
	{
		{ "command", "send_broadcast_message" },

		{ "message", message_object }
	};

	redis_client_->set(global_message_key_, "");

	return send_message(boost::json::serialize(broadcast_message), "", "");
}

auto MainServer::check_global_message()-> std::tuple<bool, std::optional<std::string>>
{
	if (thread_pool_ == nullptr)
	{
		Logger::handle().write(LogTypes::Error, "thread_pool is null");
		return { false, "thread_pool is null" };
	}

	auto job_pool = thread_pool_->job_pool();
	if (job_pool == nullptr || job_pool->lock())
	{
		Logger::handle().write(LogTypes::Error, "job_pool is null");
		return { false, "job_pool is null" };
	}

	std::this_thread::sleep_for(std::chrono::milliseconds(100));
	
	job_pool->push(
		std::make_shared<Job>(JobPriorities::High, std::bind(&MainServer::check_global_message, this), "check_global_message"));

	return consume_message_queue();
}

auto MainServer::request_client_status_update(const std::string& id, const std::string& sub_id, const std::string& message) -> std::tuple<bool, std::optional<std::string>>
{
	if (server_ == nullptr)
	{
		Logger::handle().write(LogTypes::Error, "server is null");
		return { false, "server is null" };
	}

	boost::json::object received_message = boost::json::parse(message).as_object();

	// JSON validation

	Logger::handle().write(LogTypes::Information, fmt::format("Received message: {}", message));

	redis_client_->set(id + "_" + sub_id, message, configurations_->redis_ttl_sec());

	boost::json::object message_object =
	{
		{ "message", "received connection from Server" },

		{ "command", "update_user_clinet_status" }
	};

	return send_message(boost::json::serialize(message_object), id, sub_id);
}
