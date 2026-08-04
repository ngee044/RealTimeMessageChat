#include "MainServer.h"
#include "UserClientManager.h"

#include "Logger.h"
#include "ClientHeader.h"
#include "ThreadWorker.h"
#include "JobPool.h"
#include "Job.h"
#include "JobPriorities.h"

#include <format>

#include "boost/json.hpp"
#include "boost/json/parse.hpp"

#include <vector>
#include <filesystem>
#include <algorithm>
#include <cctype>
#include <chrono>
#include <thread>
#include <expected>

using namespace Utilities;

namespace
{
	constexpr size_t max_client_id_length = 64;

	constexpr auto global_message_poll_interval = std::chrono::milliseconds(100);

	constexpr int max_drain_per_poll = 64;

	auto is_valid_client_id(const std::string& id) -> bool
	{
		if (id.empty() || id.size() > max_client_id_length)
		{
			return false;
		}

		return std::all_of(id.begin(), id.end(), [](unsigned char character) -> bool
						   { return std::isalnum(character) != 0 || character == '_' || character == '-'; });
	}
}

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
	destroy_thread_pool();

	if (server_ != nullptr)
	{
		server_->stop();
		server_.reset();
	}

	if (redis_client_ != nullptr)
	{
		redis_client_.reset();
	}
}

auto MainServer::start() -> std::expected<void, std::string>
{
	auto create_result = create_thread_pool();
	if (!create_result)
	{
		Logger::handle().write(LogTypes::Error, std::format("Failed to create thread pool: {}", create_result.error()));
		return std::unexpected(std::format("Failed to create thread pool: {}", create_result.error()));
	}

	if (configurations_->use_redis())
	{
		TLSOptions tls_options;
		tls_options.use_tls(configurations_->use_redis_tls());
		tls_options.ca_cert(configurations_->ca_cert());
		tls_options.client_cert(configurations_->client_cert());
		tls_options.client_key(configurations_->client_key());

		redis_client_ = std::make_shared<RedisClient>(configurations_->redis_host(), configurations_->redis_port(), tls_options, configurations_->redis_db_global_message_index());
		
		auto connect_result = redis_client_->connect();
		if (!connect_result)
		{
			destroy_thread_pool();
			redis_client_.reset();

			Logger::handle().write(LogTypes::Error, std::format("Failed to connect redis: {}", connect_result.error()));
			return std::unexpected(std::format("Failed to connect redis: {}", connect_result.error()));
		}

		clear_legacy_global_message_key();
	}

	auto server_result = server_->start(configurations_->server_port(), configurations_->buffer_size());
	if (!server_result)
	{
		Logger::handle().write(LogTypes::Error, std::format("Failed to start server: {}", server_result.error()));
		return std::unexpected(std::format("Failed to start server: {}", server_result.error()));
	}

	auto consume_result = thread_pool_->push(
			std::make_shared<Job>(JobPriorities::LongTerm, std::bind(&MainServer::check_global_message, this), "check_global_message"));

	if (!consume_result)
	{
		Logger::handle().write(LogTypes::Error, std::format("Failed to start consume global message job: {}", consume_result.error()));
		return std::unexpected(std::format("Failed to start consume global message job: {}", consume_result.error()));
	}
	
	return {};
}

auto MainServer::stop() -> void
{
	stopping_.store(true);

	if (server_ == nullptr)
	{
		return;
	}

	server_->stop();
}

auto MainServer::wait_stop() -> std::expected<void, std::string>
{
	if (server_ == nullptr)
	{
		Logger::handle().write(LogTypes::Error, "server is null");
		return std::unexpected("server is null");
	}

	return server_->wait_stop();
}

auto MainServer::create_thread_pool() -> std::expected<void, std::string>
{
	destroy_thread_pool();

	try
	{	
		thread_pool_ = std::make_shared<ThreadPool>();
	}
	catch(const std::bad_alloc& e)
	{
		return std::unexpected(std::format("Memory allocation failed to ThreadPool: {}", e.what()));
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
			return std::unexpected(std::format("Memory allocation failed to ThreadWorker: {}", e.what()));
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
			return std::unexpected(std::format("Memory allocation failed to ThreadWorker: {}", e.what()));
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
			return std::unexpected(std::format("Memory allocation failed to ThreadWorker: {}", e.what()));
		}

		thread_pool_->push(worker);
	}

	try
	{
		thread_pool_->push(std::make_shared<ThreadWorker>(std::vector<JobPriorities>{ JobPriorities::LongTerm }));
	}
	catch(const std::bad_alloc& e)
	{
		return std::unexpected(std::format("Memory allocation failed to ThreadWorker: {}", e.what()));
	}

	auto start_result = thread_pool_->start();
	if (!start_result)
	{
		Logger::handle().write(LogTypes::Error, std::format("Failed to start thread pool: {}", start_result.error()));
		return std::unexpected(start_result.error());
	}

	return {};
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

auto MainServer::received_connection(const std::string& id, const std::string& sub_id, bool condition) -> std::expected<void, std::string>
{
	if (server_ == nullptr)
	{
		Logger::handle().write(LogTypes::Error, "server is null");
		return std::unexpected("server is null");
	}

	if (thread_pool_ == nullptr)
	{
		Logger::handle().write(LogTypes::Error, "thread_pool is null");
		return std::unexpected("thread_pool is null");
	}

	if (condition)
	{
		if (!is_valid_client_id(id))
		{
			Logger::handle().write(LogTypes::Error, std::format("Rejected connection: invalid client id[{}]", id));
			return std::unexpected("invalid client id");
		}

		Logger::handle().write(LogTypes::Information, std::format("Received connection[{}, {}]: connected", id, sub_id));

		UserClientManager::handle().add(id, sub_id);
		return {};
	}

	Logger::handle().write(LogTypes::Information, std::format("Received connection[{}, {}]: disconnected", id, sub_id));
	
	UserClientManager::handle().remove(id, sub_id);
	return {};
}

auto MainServer::received_message(const std::string& id, const std::string& sub_id, const std::string& message) -> std::expected<void, std::string>
{
	if (server_ == nullptr)
	{
		Logger::handle().write(LogTypes::Error, "server is null");
		return std::unexpected("server is null");
	}

	if (thread_pool_ == nullptr)
	{
		Logger::handle().write(LogTypes::Error, "thread_pool is null");
		return std::unexpected("thread_pool is null");
	}

	if (message.empty())
	{
		Logger::handle().write(LogTypes::Error, "message is empty");
		return std::unexpected("message is empty");
	}

	Logger::handle().write(LogTypes::Information, std::format("Received message[{}, {}]: {}", id, sub_id, message));

	return thread_pool_->push(
		std::dynamic_pointer_cast<Job>(
			std::make_shared<ClientMessageParsing>(
				id, sub_id, message, std::bind(&MainServer::parsing_message, this, std::placeholders::_1, std::placeholders::_2, std::placeholders::_3, std::placeholders::_4)
			)
		)
	);
}

auto MainServer::send_message(const std::string& message, const std::string& id, const std::string& sub_id) -> std::expected<void, std::string>
{
	if (server_ == nullptr)
	{
		Logger::handle().write(LogTypes::Error, "server is null");
		return std::unexpected("server is null");
	}

	Logger::handle().write(LogTypes::Information, std::format("Send message[{}, {}]: {}", id, sub_id, message));

	return server_->send_message(message, id, sub_id);
}

auto MainServer::parsing_message(const std::string& id, const std::string& sub_id, const std::string& command, const std::string& message) -> std::expected<void, std::string>
{
	if (command.empty())
	{
		Logger::handle().write(LogTypes::Error, "command is empty");
		return std::unexpected("command is empty");
	}

	if (message.empty())
	{
		Logger::handle().write(LogTypes::Error, "message is empty");
		return std::unexpected("message is empty");
	}

	if (thread_pool_ == nullptr)
	{
		Logger::handle().write(LogTypes::Error, "thread_pool is null");
		return std::unexpected("thread_pool is null");
	}

	auto iter = messages_.find(command);
	if (iter == messages_.end())
	{
		Logger::handle().write(LogTypes::Error, std::format("command is not found: {}", command));
		return std::unexpected("command is not found");
	}

	return thread_pool_->push(
		std::dynamic_pointer_cast<Job>(
			std::make_shared<ClientMessageExecute>(
				id, sub_id, message, iter->second
			)
		)
	);
}

auto MainServer::consume_message_queue() -> std::expected<void, std::string>
{
	if (!configurations_->use_redis())
	{
		return {};
	}

	if (redis_client_ == nullptr)
	{
		Logger::handle().write(LogTypes::Error, "redis_client is null");
		return std::unexpected("redis_client is null");
	}

	{
		std::scoped_lock<std::mutex> lock(mutex_);
		if (!redis_client_->is_connected())
		{
			Logger::handle().write(LogTypes::Information, "Redis disconnected, attempting reconnection...");

			auto reconnect_result = redis_client_->connect();
			if (!reconnect_result)
			{
				Logger::handle().write(LogTypes::Error,
					std::format("Redis reconnection failed: {}", reconnect_result.error()));
				return std::unexpected("Redis reconnection failed");
			}

			Logger::handle().write(LogTypes::Information, "Redis reconnection successful");
		}
	}

	for (int drained = 0; drained < max_drain_per_poll; ++drained)
	{
		std::expected<std::optional<std::string>, std::string> pop_result;
		{
			std::scoped_lock<std::mutex> lock(mutex_);
			pop_result = redis_client_->lpop(global_message_key_);
		}

		if (!pop_result)
		{
			Logger::handle().write(LogTypes::Error, std::format("Failed to pop global message: {}", pop_result.error()));
			return std::unexpected(pop_result.error());
		}

		if (!pop_result.value().has_value())
		{
			return {};
		}

		const auto& payload = pop_result.value().value();

		boost::system::error_code parse_error;
		auto message_value = boost::json::parse(payload, parse_error);
		if (parse_error.failed() || !message_value.is_object())
		{
			Logger::handle().write(LogTypes::Error, std::format("Dropping global message - not a JSON object: {}", payload));
			continue;
		}

		auto received_message = message_value.as_object();
		if (!received_message.contains("id") || !received_message.at("id").is_string()
			|| !received_message.contains("sub_id") || !received_message.at("sub_id").is_string()
			|| !received_message.contains("message") || !received_message.at("message").is_object())
		{
			Logger::handle().write(LogTypes::Error, std::format("Dropping global message - schema mismatch: {}", payload));
			continue;
		}

		auto inner_message = received_message.at("message").as_object();
		if (!inner_message.contains("content") || !inner_message.at("content").is_string())
		{
			Logger::handle().write(LogTypes::Error, std::format("Dropping global message - missing message.content: {}", payload));
			continue;
		}

		boost::json::object message_object =
		{
			{ "id", received_message.at("id").as_string() },
			{ "sub_id", received_message.at("sub_id").as_string() },

			{ "data", inner_message.at("content").as_string() }
		};

		boost::json::object broadcast_message =
		{
			{ "command", "send_broadcast_message" },

			{ "message", message_object }
		};

		auto send_result = send_message(boost::json::serialize(broadcast_message), "", "");
		if (!send_result)
		{
			Logger::handle().write(LogTypes::Error,
				std::format("Broadcast failed, message dropped: {}", send_result.error()));
			return std::unexpected(send_result.error());
		}
	}

	return {};
}

auto MainServer::clear_legacy_global_message_key() -> void
{
	if (redis_client_ == nullptr)
	{
		return;
	}

	std::scoped_lock<std::mutex> lock(mutex_);

	auto length_result = redis_client_->llen(global_message_key_);
	if (length_result)
	{
		if (length_result.value() > 0)
		{
			Logger::handle().write(LogTypes::Information, std::format("Global message queue has {} pending item(s)", length_result.value()));
		}

		return;
	}

	if (length_result.error().find("WRONGTYPE") == std::string::npos)
	{
		Logger::handle().write(LogTypes::Information,
			std::format("Cannot inspect global message key: {}", length_result.error()));
		return;
	}

	Logger::handle().write(LogTypes::Information, "Removing legacy string value at global message key");
	redis_client_->del(global_message_key_);
}

auto MainServer::check_global_message()-> std::expected<void, std::string>
{
	while (!stopping_.load())
	{
		auto consume_result = consume_message_queue();
		if (!consume_result)
		{
			Logger::handle().write(LogTypes::Sequence,
				std::format("Global message poll failed, retrying: {}", consume_result.error()));
		}

		std::this_thread::sleep_for(global_message_poll_interval);
	}

	Logger::handle().write(LogTypes::Information, "Global message poll loop stopped");
	return {};
}

auto MainServer::request_client_status_update(const std::string& id, const std::string& sub_id, const std::string& message) -> std::expected<void, std::string>
{
	if (server_ == nullptr)
	{
		Logger::handle().write(LogTypes::Error, "server is null");
		return std::unexpected("server is null");
	}

	// JSON parsing with exception handling
	boost::json::object received_message;
	try
	{
		auto parsed = boost::json::parse(message);
		if (!parsed.is_object())
		{
			Logger::handle().write(LogTypes::Error, std::format("Message is not a JSON object: {}", message));
			return std::unexpected("Message is not a JSON object");
		}
		received_message = parsed.as_object();
	}
	catch (const std::exception& e)
	{
		Logger::handle().write(LogTypes::Error, std::format("JSON parsing failed: {}", e.what()));
		return std::unexpected(std::format("JSON parsing failed: {}", e.what()));
	}

	Logger::handle().write(LogTypes::Information, std::format("Received message: {}", message));

	// Null pointer check for redis_client_
	if (redis_client_ == nullptr)
	{
		Logger::handle().write(LogTypes::Information, "Redis client is not initialized, skipping status update to Redis");
	}
	else
	{
		std::scoped_lock<std::mutex> lock(mutex_);
		redis_client_->set(id + "::" + sub_id, message, configurations_->redis_ttl_sec());
	}

	boost::json::object message_object =
	{
		{ "message", "received connection from Server" },

		{ "command", "update_user_clinet_status" }
	};

	return send_message(boost::json::serialize(message_object), id, sub_id);
}
