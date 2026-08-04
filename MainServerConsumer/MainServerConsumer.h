#pragma once

#include "Configurations.h"
#include "WorkQueueConsume.h"
#include "SSLOptions.h"
#include "RedisClient.h"
#include "PostgresDB.h"

#include <string>
#include <memory>
#include <atomic>
#include <tuple>
#include <optional>
#include <expected>

using namespace RabbitMQ;
using namespace Redis;
using namespace Thread;

class MainServerConsumer
{
public:
	MainServerConsumer(std::shared_ptr<Configurations> configurations);
	~MainServerConsumer();

	auto start() -> std::expected<void, std::string>;
	auto wait_stop() -> std::expected<void, std::string>;
	auto stop() -> void;

protected:
	auto consume_queue() -> std::expected<void, std::string>;
	auto create_thread_pool() -> std::expected<void, std::string>;
	auto destroy_thread_pool() -> void;

private:
	std::atomic_bool stopped_{ false };

	std::shared_ptr<WorkQueueConsume> work_queue_consume_;
	std::shared_ptr<Configurations> configurations_;

	std::shared_ptr<ThreadPool> thread_pool_;

	const int work_queue_channel_id_;
	std::shared_ptr<RedisClient> redis_client_;
	std::shared_ptr<Database::PostgresDB> db_client_;

};
