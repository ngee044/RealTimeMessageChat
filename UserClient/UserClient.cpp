#include "UserClient.h"

#include "Converter.h"
#include "Job.h"
#include "Logger.h"
#include "JobPool.h"
#include "SystemInformation.h"
#include "ServerHeader.h"
#include "ThreadWorker.h"

#include <format>

#include "boost/json.hpp"
#include "boost/json/parse.hpp"

#include <vector>
#include <filesystem>
#include <chrono>
#include <thread>
#include <expected>

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
	stopping_.store(true);

	if (client_ != nullptr)
	{
		client_->stop();
	}

	destroy_thread_pool();

	client_.reset();
}

auto UserClient::start() -> std::expected<void, std::string>
{
	if (!client_->start(configurations_->server_ip(), configurations_->server_port(), configurations_->buffer_size()))
	{
		return std::unexpected("Failed to start client");
	}

	auto created = create_thread_pool();
	if (!created)
	{
		return std::unexpected(std::format("Failed to create thread pool: {}", created.error()));
	}

	client_->wait_stop();

	// Properly stop thread pool after client stops
	destroy_thread_pool();

	return {};
}

auto UserClient::stop() -> void
{
	stopping_.store(true);

	if (client_ == nullptr)
	{
		return;
	}

	client_->stop();
}

auto UserClient::create_thread_pool() -> std::expected<void, std::string>
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
			return std::unexpected(std::format("Memory allocation failed to ThreadWorker: {}", e.what()));
		}

		thread_pool_->push(worker);
	}

	auto started = thread_pool_->start();
	if (!started)
	{
		Logger::handle().write(LogTypes::Error, std::format("Failed to start thread pool: {}", started.error()));
		return std::unexpected(started.error());
	}

	return {};
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

auto UserClient::received_connection(bool condition, bool by_itself) -> std::expected<void, std::string>
{
	if (client_ == nullptr)
	{
		return std::unexpected("client is null");
	}

	Logger::handle().write(LogTypes::Information, std::format("received condition message from Server : {}", condition));

	if (!condition)
	{
		// Connection lost - attempt reconnection if not stopped by itself
		if (!by_itself)
		{
			Logger::handle().write(LogTypes::Information, "Connection lost, attempting reconnection...");

			// Schedule reconnection attempt in thread pool
			if (thread_pool_ != nullptr)
			{
				constexpr int reconnect_delay_sec = 5;
				constexpr int max_reconnect_attempts = 10;

				thread_pool_->push(
					std::make_shared<Job>(JobPriorities::High,
						[this, reconnect_delay_sec, max_reconnect_attempts]() -> std::expected<void, std::string>
						{
							for (int attempt = 0; attempt < max_reconnect_attempts; ++attempt)
							{
								constexpr auto poll_interval = std::chrono::milliseconds(100);
								const int poll_count = reconnect_delay_sec * 1000 / static_cast<int>(poll_interval.count());

								for (int elapsed = 0; elapsed < poll_count; ++elapsed)
								{
									if (stopping_.load())
									{
										return {};
									}

									std::this_thread::sleep_for(poll_interval);
								}

								if (stopping_.load())
								{
									return {};
								}

								Logger::handle().write(LogTypes::Information,
									std::format("Reconnection attempt {}/{}", attempt + 1, max_reconnect_attempts));

								if (client_ != nullptr && client_->start(configurations_->server_ip(),
									configurations_->server_port(), configurations_->buffer_size()))
								{
									Logger::handle().write(LogTypes::Information, "Reconnection successful");
									return {};
								}
							}

							Logger::handle().write(LogTypes::Error,
								std::format("Failed to reconnect after {} attempts", max_reconnect_attempts));
							return std::unexpected("Failed to reconnect");
						}, "reconnect_job"));
			}
		}

		return std::unexpected("Connection lost");
	}

	auto job_pool = thread_pool_->job_pool();
	if (job_pool == nullptr)
	{
		return std::unexpected("job_pool is null");
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

auto UserClient::received_message(const std::string& message) -> std::expected<void, std::string>
{
	if (client_ == nullptr)
	{
		return std::unexpected("client is null");
	}

	if (thread_pool_ == nullptr)
	{
		return std::unexpected("thread_pool is null");
	}

	return thread_pool_->push(
			std::dynamic_pointer_cast<Job>(
				std::make_shared<ServerMessageParsing>(
					message, std::bind(&UserClient::parsing_message, this, std::placeholders::_1, std::placeholders::_2)
				)
			)
		);

}

auto UserClient::parsing_message(const std::string& command, const std::string& message) -> std::expected<void, std::string>
{
	if (client_ == nullptr)
	{
		return std::unexpected("client is null");
	}

	if (thread_pool_ == nullptr)
	{
		return std::unexpected("thread_pool is null");
	}

	if (command.empty())
	{
		return std::unexpected("command is empty");
	}

	if (message.empty())
	{
		return std::unexpected("message is empty");
	}

	auto iter = messages_.find(command);
	if (iter == messages_.end())
	{
		Logger::handle().write(LogTypes::Error, std::format("command is not found: {}", command));
		return std::unexpected("command is not found");
	}

	return thread_pool_->push(
			std::dynamic_pointer_cast<Job>(
				std::make_shared<ServerMessageExecute>(
					message, iter->second
				)
			)
		);
}

auto UserClient::update_user_clinet_status(const std::string message) -> std::expected<void, std::string>
{
	Logger::handle().write(LogTypes::Information, std::format("Received client status update: {}", message));

	return {};
}

auto UserClient::send_broadcast_message(const std::string message) -> std::expected<void, std::string>
{
	Logger::handle().write(LogTypes::Information, std::format("Received broadcast message: {}", message));

	return {};
}
