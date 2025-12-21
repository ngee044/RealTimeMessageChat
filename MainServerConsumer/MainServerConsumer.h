#pragma once

#include "Configurations.h"
#include "WorkQueueConsume.h"
#include "SSLOptions.h"
#include "RedisClient.h"
#include "PostgresDB.h"

#include <string>
#include <memory>
#include <tuple>
#include <optional>

using namespace RabbitMQ;
using namespace Redis;
using namespace Thread;

class MainServerConsumer
{
public:
	MainServerConsumer(std::shared_ptr<Configurations> configurations);
	~MainServerConsumer();

	auto start() -> std::tuple<bool, std::optional<std::string>>;
	auto wait_stop() -> std::tuple<bool, std::optional<std::string>>;
	auto stop() -> void;

protected:
	auto consume_queue() -> std::tuple<bool, std::optional<std::string>>;
	auto create_thread_pool() -> std::tuple<bool, std::optional<std::string>>;
	auto destroy_thread_pool() -> void;

private:
	std::shared_ptr<WorkQueueConsume> work_queue_consume_;
	std::shared_ptr<Configurations> configurations_;

	std::shared_ptr<ThreadPool> thread_pool_;

	const int work_queue_channel_id_;
	std::shared_ptr<RedisClient> redis_client_;
	std::shared_ptr<Database::PostgresDB> db_client_;

};
