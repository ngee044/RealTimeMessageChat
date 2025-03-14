#include "MainServerConsumer.h"

#include "Logger.h"
#include "Converter.h"
#include "ThreadWorker.h"

#include "fmt/xchar.h"
#include "fmt/format.h"

#include "boost/json.hpp"
#include "boost/json/parse.hpp"

#include <filesystem>
#include <iostream>

using namespace Utilities;

MainServerConsumer::MainServerConsumer(std::shared_ptr<Configurations> configurations)
	: configurations_(configurations)
	, work_queue_consume_(nullptr)
	, work_queue_channel_id_(1)
	, redis_client_(nullptr)
{
}

MainServerConsumer::~MainServerConsumer()
{
	stop();
}

auto MainServerConsumer::create_thread_pool() -> std::tuple<bool, std::optional<std::string>>
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

auto MainServerConsumer::destroy_thread_pool() -> void
{
	if (thread_pool_ == nullptr)
	{
		return;
	}

	thread_pool_->stop();
	thread_pool_.reset();
}


auto MainServerConsumer::start() -> std::tuple<bool, std::optional<std::string>>
{	
	auto [result, error_message] = create_thread_pool();
	if (!result)
	{
		Logger::handle().write(LogTypes::Error, fmt::format("Failed to create thread pool: {}", error_message.value()));
		return { false, fmt::format("Failed to create thread pool: {}", error_message.value()) };
	}

	SSLOptions ssl_options;
	ssl_options.use_ssl(configurations_->use_ssl());
	ssl_options.ca_cert(configurations_->ca_cert());
	ssl_options.engine(configurations_->engine());
	ssl_options.client_cert(configurations_->client_cert());
	ssl_options.client_key(configurations_->client_key());

	work_queue_consume_ = std::make_shared<WorkQueueConsume>(configurations_->rabbit_mq_host(), 
															 configurations_->rabbit_mq_port(), 
															 configurations_->rabbit_mq_user_name(),
															 configurations_->rabbit_mq_password(), ssl_options);
														 
	std::tie(result, error_message) = work_queue_consume_->start();
	if (!result)
	{
		Logger::handle().write(LogTypes::Error, fmt::format("Failed to start work queue consume: {}", error_message.value()));
		return { false, fmt::format("Failed to start work queue consume: {}", error_message.value()) };
	}
	Logger::handle().write(LogTypes::Information, "work queue consume started");

	std::tie(result, error_message) = work_queue_consume_->connect(60);
	if (!result)
	{
		Logger::handle().write(LogTypes::Error, fmt::format("Failed to connect work queue consume: {}", error_message.value()));
		return { false, fmt::format("Failed to connect work queue consume: {}", error_message.value()) };
	}
	Logger::handle().write(LogTypes::Information, "work queue consume connected");

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
			work_queue_consume_->stop();
			work_queue_consume_.reset();

			redis_client_.reset();

			Logger::handle().write(LogTypes::Error, fmt::format("Failed to connect redis: {}", connect_error.value()));
			return { false, fmt::format("Failed to connect redis: {}", connect_error.value()) };
		}
	}
	else
	{
		Logger::handle().write(LogTypes::Error, "Redis is not used");
		return { false, "Redis is not used" };
	}

	Logger::handle().write(LogTypes::Information, "redis connected");

	std::tie(result, error_message) = consume_queue();
	if (!result)
	{
		destroy_thread_pool();
		work_queue_consume_->stop();
		work_queue_consume_.reset();

		redis_client_.reset();

		Logger::handle().write(LogTypes::Error, fmt::format("Failed to consume queue: {}", error_message.value()));
		return { false, fmt::format("Failed to consume queue: {}", error_message.value()) };
	}

	return { true, std::nullopt };
}

auto MainServerConsumer::wait_stop() -> std::tuple<bool, std::optional<std::string>>
{
	if (work_queue_consume_ == nullptr)
	{
		Logger::handle().write(LogTypes::Error, "work_queue_consume is null");
		return { false, "work_queue_consume is null" };
	}

	work_queue_consume_->wait_stop();

	return { true, std::nullopt };
}

auto MainServerConsumer::stop() -> void
{
	if (work_queue_consume_ != nullptr)
	{
		work_queue_consume_->stop();
		work_queue_consume_.reset();
	}

	destroy_thread_pool();

	if (redis_client_ != nullptr)
	{
		redis_client_->disconnect();
		redis_client_.reset();
	}
}

auto MainServerConsumer::consume_queue() -> std::tuple<bool, std::optional<std::string>>
{
	if (configurations_ == nullptr)
	{
		Logger::handle().write(LogTypes::Error, "configurations is null");
		return { false, "configurations is null" };
	}

	if (work_queue_consume_ == nullptr)
	{
		Logger::handle().write(LogTypes::Error, "work_queue_consume is null");
		return { false, "work_queue_consume is null" };
	}

	if (redis_client_ == nullptr)
	{
		Logger::handle().write(LogTypes::Error, "redis_client is null");
		return { false, "redis_client is null" };
	}

	auto [declred_name, error] = work_queue_consume_->channel_open(work_queue_channel_id_, configurations_->consume_queue_name());
	if (!declred_name.has_value())
	{
		Logger::handle().write(LogTypes::Error, fmt::format("Failed to open channel: {}", error.value()));
		return { false, error };
	}

	auto [prepare_success, prepare_error] = work_queue_consume_->prepare_consume();
	if (!prepare_success)
	{
		return { false, fmt::format("cannot prepare consume: {}", prepare_error.value()) };
	}

	auto [consume_register, consume_error] = work_queue_consume_->register_consume(work_queue_channel_id_, configurations_->consume_queue_name(), 
		[&](const std::string& queue_name, const std::string& message, const std::string& message_type)-> std::tuple<bool, std::optional<std::string>>
		{
			// TODO 
			// log type: sequence
			Logger::handle().write(LogTypes::Information, fmt::format("consume message: queue_name[{}] => {}", queue_name, message));

			auto message_value = boost::json::parse(message);
			if (!message_value.is_object())
			{
				Logger::handle().write(LogTypes::Error, fmt::format("Failed to parse message: {}", message));
				return { false, "Failed to parse message" };
			}

			auto received_message = message_value.as_object();
			if (!received_message.contains("id") || !received_message.at("id").is_string())
			{
				Logger::handle().write(LogTypes::Error, fmt::format("Failed to parse message: {}", message));
				return { false, "Failed to parse message" };
			}

			if (!received_message.contains("sub_id") || !received_message.at("sub_id").is_string())
			{
				Logger::handle().write(LogTypes::Error, fmt::format("Failed to parse message: {}", message));
				return { false, "Failed to parse message" };
			}

			if (!received_message.contains("message") || !received_message.at("message").is_string())
			{
				Logger::handle().write(LogTypes::Error, fmt::format("Failed to parse message: {}", message));
				return { false, "Failed to parse message" };
			}

			redis_client_->set(configurations_->global_message_key(), message);

			return { true, std::nullopt };
		});

	if (!consume_register)
	{
		Logger::handle().write(LogTypes::Error, fmt::format("Failed to start consume: {}", consume_error.value()));
		return { false, consume_error };
	}

	auto [consume_start, consume_start_error] = work_queue_consume_->start_consume();
	if (!consume_start)
	{
		Logger::handle().write(LogTypes::Error, fmt::format("Failed to start consume: {}", consume_start_error.value()));
		return { false, consume_start_error };
	}

	return { true, std::nullopt };	
}

