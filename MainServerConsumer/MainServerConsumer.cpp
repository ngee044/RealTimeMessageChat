#include "MainServerConsumer.h"
#include "DBWorker.h"

#include "Logger.h"
#include "Converter.h"
#include "ThreadWorker.h"

#include <format>

#include "boost/json.hpp"
#include "boost/json/parse.hpp"

#include <filesystem>
#include <iostream>
#include <expected>

using namespace Utilities;

MainServerConsumer::MainServerConsumer(std::shared_ptr<Configurations> configurations)
	: configurations_(configurations)
	, work_queue_consume_(nullptr)
	, work_queue_channel_id_(1)
	, redis_client_(nullptr)
	, db_client_(nullptr)
{
}

MainServerConsumer::~MainServerConsumer()
{
	stop();
}

auto MainServerConsumer::create_thread_pool() -> std::expected<void, std::string>
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

	auto start_result = thread_pool_->start();
	if (!start_result)
	{
		Logger::handle().write(LogTypes::Error, std::format("Failed to start thread pool: {}", start_result.error()));
		return std::unexpected(start_result.error());
	}

	return {};
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


auto MainServerConsumer::start() -> std::expected<void, std::string>
{	
	auto thread_pool_result = create_thread_pool();
	if (!thread_pool_result)
	{
		Logger::handle().write(LogTypes::Error, std::format("Failed to create thread pool: {}", thread_pool_result.error()));
		return std::unexpected(std::format("Failed to create thread pool: {}", thread_pool_result.error()));
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
														 
	auto queue_start_result = work_queue_consume_->start();
	if (!queue_start_result)
	{
		Logger::handle().write(LogTypes::Error, std::format("Failed to start work queue consume: {}", queue_start_result.error()));
		return std::unexpected(std::format("Failed to start work queue consume: {}", queue_start_result.error()));
	}
	Logger::handle().write(LogTypes::Information, "work queue consume started");

	// Connect with retry logic
	constexpr int max_retries = 5;
	constexpr int retry_delay_sec = 5;
	bool connected = false;
	std::string connect_error_message = "Unknown error";

	for (int retry = 0; retry < max_retries; ++retry)
	{
		auto queue_connect_result = work_queue_consume_->connect(60);
		if (queue_connect_result)
		{
			connected = true;
			break;
		}

		connect_error_message = queue_connect_result.error();

		if (retry < max_retries - 1)
		{
			Logger::handle().write(LogTypes::Information,
				std::format("RabbitMQ connection failed, retry {}/{} in {} seconds...", retry + 1, max_retries, retry_delay_sec));
			std::this_thread::sleep_for(std::chrono::seconds(retry_delay_sec));
		}
	}

	if (!connected)
	{
		Logger::handle().write(LogTypes::Error, std::format("Failed to connect work queue consume after {} retries: {}", max_retries, connect_error_message));
		return std::unexpected(std::format("Failed to connect work queue consume after {} retries", max_retries));
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

		auto redis_connect_result = redis_client_->connect();
		if (!redis_connect_result)
		{
			destroy_thread_pool();
			work_queue_consume_->stop();
			work_queue_consume_.reset();

			redis_client_.reset();

			Logger::handle().write(LogTypes::Error, std::format("Failed to connect redis: {}", redis_connect_result.error()));
			return std::unexpected(std::format("Failed to connect redis: {}", redis_connect_result.error()));
		}
	}
	else
	{
		Logger::handle().write(LogTypes::Error, "Redis is not used");
		return std::unexpected("Redis is not used");
	}

	Logger::handle().write(LogTypes::Information, "redis connected");

	// Initialize database connection if enabled
	if (configurations_->use_database())
	{
		std::string db_error;
		try
		{
			db_client_ = std::make_shared<Database::PostgresDB>(configurations_->database_connection_string());

			if (db_client_->handler() == nullptr || PQstatus(db_client_->handler()) != CONNECTION_OK)
			{
				db_error = "PostgreSQL connection is not established";
			}
		}
		catch (const std::exception& e)
		{
			db_error = e.what();
		}

		if (!db_error.empty())
		{
			destroy_thread_pool();
			work_queue_consume_->stop();
			work_queue_consume_.reset();
			redis_client_.reset();
			db_client_.reset();

			Logger::handle().write(LogTypes::Error, std::format("Failed to initialize database: {}", db_error));
			return std::unexpected(std::format("Failed to initialize database: {}", db_error));
		}

		Logger::handle().write(LogTypes::Information, "Database client initialized successfully");
	}
	else
	{
		Logger::handle().write(LogTypes::Information, "Database is not enabled");
	}

	auto consume_result = consume_queue();
	if (!consume_result)
	{
		destroy_thread_pool();
		work_queue_consume_->stop();
		work_queue_consume_.reset();

		redis_client_.reset();

		Logger::handle().write(LogTypes::Error, std::format("Failed to consume queue: {}", consume_result.error()));
		return std::unexpected(std::format("Failed to consume queue: {}", consume_result.error()));
	}

	return {};
}

auto MainServerConsumer::wait_stop() -> std::expected<void, std::string>
{
	if (work_queue_consume_ == nullptr)
	{
		Logger::handle().write(LogTypes::Error, "work_queue_consume is null");
		return std::unexpected("work_queue_consume is null");
	}

	work_queue_consume_->wait_stop();

	return {};
}

auto MainServerConsumer::stop() -> void
{
	if (stopped_.exchange(true))
	{
		return;
	}

	if (work_queue_consume_ != nullptr)
	{
		constexpr int max_stop_consume_attempts = 3;
		bool consume_stopped = false;
		std::string stop_consume_error = "Unknown error";

		for (int attempt = 0; attempt < max_stop_consume_attempts; ++attempt)
		{
			auto stop_consume_result = work_queue_consume_->stop_consume();
			if (stop_consume_result)
			{
				consume_stopped = true;
				break;
			}

			stop_consume_error = stop_consume_result.error();

			std::this_thread::sleep_for(std::chrono::milliseconds(50));
		}

		if (!consume_stopped)
		{
			Logger::handle().write(LogTypes::Error,
				std::format("Failed to stop consume after {} attempts: {}", max_stop_consume_attempts,
							stop_consume_error));
		}

		work_queue_consume_->stop();

	}

	destroy_thread_pool();

	if (redis_client_ != nullptr)
	{
		redis_client_->disconnect();
		redis_client_.reset();
	}

	if (db_client_ != nullptr)
	{
		db_client_.reset();
	}
}

auto MainServerConsumer::consume_queue() -> std::expected<void, std::string>
{
	if (configurations_ == nullptr)
	{
		Logger::handle().write(LogTypes::Error, "configurations is null");
		return std::unexpected("configurations is null");
	}

	if (work_queue_consume_ == nullptr)
	{
		Logger::handle().write(LogTypes::Error, "work_queue_consume is null");
		return std::unexpected("work_queue_consume is null");
	}

	if (redis_client_ == nullptr)
	{
		Logger::handle().write(LogTypes::Error, "redis_client is null");
		return std::unexpected("redis_client is null");
	}

	auto declred_name = work_queue_consume_->channel_open(work_queue_channel_id_, configurations_->consume_queue_name());
	if (!declred_name)
	{
		Logger::handle().write(LogTypes::Error, std::format("Failed to open channel: {}", declred_name.error()));
		return std::unexpected(declred_name.error());
	}

	auto prepare_result = work_queue_consume_->prepare_consume();
	if (!prepare_result)
	{
		return std::unexpected(std::format("cannot prepare consume: {}", prepare_result.error()));
	}

	auto register_result = work_queue_consume_->register_consume(work_queue_channel_id_, configurations_->consume_queue_name(),
		[&](const std::string& queue_name, const std::string& message, const std::string& message_type)-> std::expected<void, std::string>
		{
			try
			{
				Logger::handle().write(LogTypes::Sequence, std::format("consume message: queue_name[{}] => {}", queue_name, message));

				boost::json::value message_value;
				try
				{
					message_value = boost::json::parse(message);
				}
				catch (const std::exception& e)
				{
					Logger::handle().write(LogTypes::Error, std::format("Dropping malformed message - JSON parsing failed: {}", e.what()));
					return {};
				}

				if (!message_value.is_object())
				{
					Logger::handle().write(LogTypes::Error, std::format("Dropping message - root is not a JSON object: {}", message));
					return {};
				}

				auto received_message = message_value.as_object();
				if (!received_message.contains("id") || !received_message.at("id").is_string())
				{
					Logger::handle().write(LogTypes::Error, std::format("Dropping message - missing 'id' field: {}", message));
					return {};
				}

				if (!received_message.contains("sub_id") || !received_message.at("sub_id").is_string())
				{
					Logger::handle().write(LogTypes::Error, std::format("Dropping message - missing 'sub_id' field: {}", message));
					return {};
				}

				if (!received_message.contains("message") || !received_message.at("message").is_object())
				{
					Logger::handle().write(LogTypes::Error, std::format("Dropping message - missing 'message' object: {}", message));
					return {};
				}

				auto inner_message = received_message.at("message").as_object();
				if (!inner_message.contains("content") || !inner_message.at("content").is_string())
				{
					Logger::handle().write(LogTypes::Error, std::format("Dropping message - missing 'message.content': {}", message));
					return {};
				}

				// Store message in Redis for MainServer to broadcast
				if (redis_client_ == nullptr)
				{
					Logger::handle().write(LogTypes::Error, "Redis client is null, cannot store message");
					return std::unexpected("Redis client is null");
				}

				const auto& global_key = configurations_->global_message_key();
				auto queue_result = redis_client_->rpush(global_key, { message });
				if (!queue_result)
				{
					auto length_result = redis_client_->llen(global_key);
					const bool wrong_type = !length_result && length_result.error().find("WRONGTYPE") != std::string::npos;

					if (!wrong_type)
					{
						Logger::handle().write(LogTypes::Error,
							std::format("Failed to queue message in Redis: {}", queue_result.error()));
						return std::unexpected("Failed to queue message in Redis");
					}

					Logger::handle().write(LogTypes::Information, "Removing legacy string value at global message key");
					redis_client_->del(global_key);

					auto requeue_result = redis_client_->rpush(global_key, { message });
					if (!requeue_result)
					{
						Logger::handle().write(LogTypes::Error,
							std::format("Failed to queue message in Redis after cleanup: {}", requeue_result.error()));
						return std::unexpected("Failed to queue message in Redis");
					}
				}

				// Store message in database asynchronously if database is enabled
				if (configurations_->use_database() && db_client_ != nullptr)
				{
					try
					{
						auto db_worker = std::make_shared<Database::DBWorker>(
							db_client_,
							message,
							configurations_->database_encryption_enabled(),
							configurations_->database_encryption_key(),
							configurations_->database_encryption_iv(),
							Thread::JobPriorities::Low
						);

						auto push_result = thread_pool_->push(db_worker);
						if (!push_result)
						{
							Logger::handle().write(LogTypes::Error,
								std::format("Failed to push DBWorker to thread pool: {}", push_result.error()));
							// Don't return error - database storage is non-critical
						}
						else
						{
							Logger::handle().write(LogTypes::Information, "DBWorker job pushed to thread pool");
						}
					}
					catch (const std::exception& e)
					{
						Logger::handle().write(LogTypes::Error,
							std::format("Exception creating DBWorker: {}", e.what()));
						// Don't return error - database storage is non-critical
					}
				}

				return {};
			}
			catch (const std::exception& e)
			{
				Logger::handle().write(LogTypes::Error, std::format("Consume callback exception: {}", e.what()));
				return std::unexpected(std::format("Consume callback exception: {}", e.what()));
			}
		});

	if (!register_result)
	{
		Logger::handle().write(LogTypes::Error, std::format("Failed to start consume: {}", register_result.error()));
		return std::unexpected(register_result.error());
	}

	auto consume_start_result = work_queue_consume_->start_consume();
	if (!consume_start_result)
	{
		Logger::handle().write(LogTypes::Error, std::format("Failed to start consume: {}", consume_start_result.error()));
		return std::unexpected(consume_start_result.error());
	}

	return {};	
}
