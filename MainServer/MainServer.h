#pragma once

#include "Configurations.h"

#include "NetworkServer.h"
#include "SSLOptions.h"
#include "RedisClient.h"

#include <string>
#include <memory>
#include <tuple>
#include <optional>
#include <map>
#include <functional>

using namespace Thread;
using namespace Network;
using namespace RabbitMQ;
using namespace Redis;

class MainServer
{
public:
	MainServer(std::shared_ptr<Configurations> configurations);
	~MainServer(void);

	auto start() -> std::tuple<bool, std::optional<std::string>>;
	auto stop() -> void;

protected:
	auto create_thread_pool() -> std::tuple<bool, std::optional<std::string>>;
	auto destroy_thread_pool() -> void;

	auto received_connection(const std::string& id, const std::string& sub_id, const bool& condition) -> std::tuple<bool, std::optional<std::string>>;
	auto received_message(const std::string& id, const std::string& sub_id, const std::string& message) -> std::tuple<bool, std::optional<std::string>>;
	auto send_message(const std::string& message, const std::string& id = "", const std::string& sub_id = "") -> std::tuple<bool, std::optional<std::string>>;
	auto parsing_message(const std::string& id, const std::string& sub_id, const std::string& command, const std::string& message) -> std::tuple<bool, std::optional<std::string>>;
	
	// jobs
	auto db_periodic_update_callback() -> std::tuple<bool, std::optional<std::string>>;

	// message list
	auto publish_message_queue(const std::string& id, const std::string& sub_id, const std::string& message) -> std::tuple<bool, std::optional<std::string>>;
	auto request_client_status_update(const std::string& id, const std::string& sub_id, const std::string& message) -> std::tuple<bool, std::optional<std::string>>;

private:
	std::mutex mutex_;

	std::shared_ptr<NetworkServer> server_;
	std::shared_ptr<ThreadPool> thread_pool_;
	std::shared_ptr<Configurations> configurations_;

	// If you use more than 3 Redis clients, manage them.
	std::shared_ptr<RedisClient> redis_client_;
	std::shared_ptr<RedisClient> local_redis_client_;

	std::string register_key_;

	std::map<std::string, std::function<std::tuple<bool, std::optional<std::string>>(const std::string&, const std::string&, const std::string&)>> messages_;
};