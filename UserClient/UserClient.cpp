#include "UserClient.h"

#include "Converter.h"
#include "Job.h"
#include "Logger.h"
#include "JobPool.h"
#include "SystemInformation.h"
#include "ServerHeader.h"
#include "ThreadWorker.h"

#include "fmt/format.h"
#include "fmt/xchar.h"

#include "boost/json.hpp"
#include "boost/json/parse.hpp"

#include <vector>
#include <filesystem>

UserClient::UserClient(std::shared_ptr<Configurations> configurations)
	: thread_pool_(nullptr)
	, configurations_(configurations)
	, register_key_("MainServer")
	, messages_()
{
	// title is id
	client_ = std::make_shared<NetworkClient>(configurations->client_title(), configurations->high_priority_count(), configurations->normal_priority_count(), configurations->low_priority_count());

	client_->register_key(register_key_);
	client_->received_connection_callback(std::bind(&UserClient::received_connection, this, std::placeholders::_1, std::placeholders::_2));
	client_->received_message_callback(std::bind(&UserClient::received_message, this, std::placeholders::_1));

	messages_.insert({ "update_user_clinet_status", std::bind(&UserClient::update_user_clinet_status, this, std::placeholders::_1) });
	messages_.insert({ "send_broadcast_message", std::bind(&UserClient::send_broadcast_message, this, std::placeholders::_1) });
}

UserClient::~UserClient(void)
{
	client_->stop();
	client_.reset();

	destroy_thread_pool();
}

auto UserClient::start() -> std::tuple<bool, std::optional<std::string>>
{
	if (!client_->start(configurations_->server_ip(), configurations_->server_port(), configurations_->buffer_size()))
	{
		return { false, "Failed to start client" };
	}

	auto [result, error_message] = create_thread_pool();
	if (!result)
	{
		return { false, fmt::format("Failed to create thread pool: {}", error_message.value()) };
	}

	client_->wait_stop();

	return { true, std::nullopt };
}

auto UserClient::stop() -> void
{
	client_->stop();

	destroy_thread_pool();
}

auto UserClient::create_thread_pool() -> std::tuple<bool, std::optional<std::string>>
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

		thread_pool_->push(worker );
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
		return { false, message };
	}

	return { true, std::nullopt };
}

auto UserClient::destroy_thread_pool() -> void
{
	std::scoped_lock<std::mutex> lock(mutex_);

	if (thread_pool_ == nullptr)
	{
		return;
	}

	thread_pool_->stop(true);
	thread_pool_.reset();
}

auto UserClient::received_connection(const bool& condition, const bool& by_itself) -> std::tuple<bool, std::optional<std::string>>
{
	if (client_ == nullptr)
	{
		return { false, "client is null" };
	}

	Logger::handle().write(LogTypes::Information, fmt::format("received condition message from Server : {}", condition));

	if (!condition)
	{
		client_->stop();
		return { false, "received condition message from Server" };
	}

	auto job_pool = thread_pool_->job_pool();
	if (job_pool == nullptr)
	{
		return { false, "job_pool is null" };
	}

	boost::json::object message =
	{
		{ "id", client_->id() },
		{ "sub_id", client_->sub_id() },
		{ "message", "received connection from Server" },

		{ "command", "request_client_status_update" }
	};
	
	return client_->send_message(boost::json::serialize(message));
}

auto UserClient::received_message(const std::string& message) -> std::tuple<bool, std::optional<std::string>>
{
	if (client_ == nullptr)
	{
		return { false, "client is null" };
	}

	if (thread_pool_ == nullptr)
	{
		return { false, "thread_pool is null" };
	}

	return thread_pool_->push(
			std::dynamic_pointer_cast<Job>(
				std::make_shared<ServerMessageParsing>(
					client_->id(), message, std::bind(&UserClient::parsing_message, this, std::placeholders::_1, std::placeholders::_2)
				)
			)
		);

}

auto UserClient::parsing_message(const std::string& command, const std::string& message) -> std::tuple<bool, std::optional<std::string>>
{
	if (client_ == nullptr)
	{
		return { false, "client is null" };
	}

	if (thread_pool_ == nullptr)
	{
		return { false, "thread_pool is null" };
	}

	if (command.empty())
	{
		return { false, "command is empty" };
	}

	if (message.empty())
	{
		return { false, "message is empty" };
	}

	auto iter = messages_.find(command);
	if (iter == messages_.end())
	{
		Logger::handle().write(LogTypes::Error, fmt::format("command is not found: {}", command));
		return { false, "command is not found" };
	}

	return thread_pool_->push(
			std::dynamic_pointer_cast<Job>(
				std::make_shared<ServerMessageExecute>(
					client_->id(), message, iter->second
				)
			)
		);
}

auto UserClient::update_user_clinet_status(const std::string message) -> std::tuple<bool, std::optional<std::string>>
{
	Logger::handle().write(LogTypes::Information, fmt::format("Received message: {}", message));
	
	boost::json::object send_message =
	{
		{ "id", client_->id() },
		{ "sub_id", client_->sub_id() },
		{ "message", "received connection from Server" },

		{ "command", "request_client_status_update" }
	};
	
	return client_->send_message(boost::json::serialize(send_message));
}

auto UserClient::send_broadcast_message(const std::string message) -> std::tuple<bool, std::optional<std::string>>
{
	Logger::handle().write(LogTypes::Information, fmt::format("Received broadcast message: {}", message));

	return { true, std::nullopt };
}
