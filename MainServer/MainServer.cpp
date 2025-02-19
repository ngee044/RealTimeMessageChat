#include "MainServer.h"
#include "UserClientManager.h"
#include "DBPeriodicUpdateJob.h"

#include "Logger.h"
#include "ClientHeader.h"
#include "ThreadWorker.h"

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
{
	server_ = std::make_shared<NetworkServer>(configurations->client_title(), configurations->high_priority_count(), configurations->normal_priority_count(), configurations->low_priority_count());
	
	server_->register_key(register_key_);
	server_->received_connection_callback(std::bind(&MainServer::received_connection, this, std::placeholders::_1, std::placeholders::_2, std::placeholders::_3));
	server_->received_message_callback(std::bind(&MainServer::received_message, this, std::placeholders::_1, std::placeholders::_2, std::placeholders::_3));

	messages_.insert({ "publish_message_queue", std::bind(&MainServer::publish_message_queue, this, std::placeholders::_1, std::placeholders::_2, std::placeholders::_3) });
	messages_.insert({ "request_client_status_update", std::bind(&MainServer::request_client_status_update, this, std::placeholders::_1, std::placeholders::_2, std::placeholders::_3) });

}

MainServer::~MainServer(void)
{
	stop();

	destroy_thread_pool();
}

auto MainServer::start() -> std::tuple<bool, std::optional<std::string>>
{
	auto [result, error_message] = server_->start(configurations_->server_port(), configurations_->buffer_size());
	if (!result)
	{
		Logger::handle().write(LogTypes::Error, fmt::format("Failed to start server: {}", error_message.value()));
		return { false, fmt::format("Failed to start server: {}", error_message.value()) };
	}

	std::tie(result, error_message) = create_thread_pool();
	if (!result)
	{
		Logger::handle().write(LogTypes::Error, fmt::format("Failed to create thread pool: {}", error_message.value()));
		return { false, fmt::format("Failed to create thread pool: {}", error_message.value()) };
	}

	server_->wait_stop();

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

auto MainServer::db_periodic_update_callback() -> std::tuple<bool, std::optional<std::string>>
{
	// TODO
	// update global information in PostgreSQL

	return {false, "Not implemented"};
}

auto MainServer::publish_message_queue(const std::string& id, const std::string& sub_id, const std::string& message) -> std::tuple<bool, std::optional<std::string>>
{
	if (server_ == nullptr)
	{
		Logger::handle().write(LogTypes::Error, "server is null");
		return { false, "server is null" };
	}

	boost::json::object received_message = boost::json::parse(message).as_object();

	// TODO
	// JSON validation

	Logger::handle().write(LogTypes::Information, fmt::format("Received message: {}", message));

	// TODO
	// Publish Message Queue

	return { true, std::nullopt };
}

auto MainServer::request_client_status_update(const std::string& id, const std::string& sub_id, const std::string& message) -> std::tuple<bool, std::optional<std::string>>
{
	if (server_ == nullptr)
	{
		Logger::handle().write(LogTypes::Error, "server is null");
		return { false, "server is null" };
	}

	boost::json::object received_message = boost::json::parse(message).as_object();

	// TODO
	// JSON validation

	Logger::handle().write(LogTypes::Information, fmt::format("Received message: {}", message));

	// TODO
	// Update current client information in Redis

	// write redis
	boost::json::object response_message =
	{
		{ "command", "response_client_status_update"},

		{ "message", "..." }
	};

	// TODO
	// update postgreSQL
	// this thread is longterm job
	thread_pool_->push(
		std::dynamic_pointer_cast<Job>(
			std::make_shared<DBPeriodicUpdateJob>(
				id, sub_id, boost::json::serialize(response_message), std::bind(&MainServer::db_periodic_update_callback, this)
			)
		)
	);


	return send_message(boost::json::serialize(response_message), id, sub_id);
}
