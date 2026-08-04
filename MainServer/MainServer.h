#pragma once

#include "Configurations.h"

#include "NetworkServer.h"
#include "SSLOptions.h"
#include "RedisClient.h"

#include <string>
#include <memory>
#include <tuple>
#include <optional>
#include <expected>
#include <map>
#include <functional>
#include <atomic>
#include <mutex>

using namespace Thread;
using namespace Network;
using namespace RabbitMQ;
using namespace Redis;

class MainServer
{
public:
	MainServer(std::shared_ptr<Configurations> configurations);
	~MainServer(void);

	auto start() -> std::expected<void, std::string>;
	auto stop() -> void;
	auto wait_stop() -> std::expected<void, std::string>;

protected:
	auto create_thread_pool() -> std::expected<void, std::string>;
	auto destroy_thread_pool() -> void;

	auto received_connection(const std::string& id, const std::string& sub_id, bool condition) -> std::expected<void, std::string>;
	auto received_message(const std::string& id, const std::string& sub_id, const std::string& message) -> std::expected<void, std::string>;
	auto send_message(const std::string& message, const std::string& id = "", const std::string& sub_id = "") -> std::expected<void, std::string>;
	auto parsing_message(const std::string& id, const std::string& sub_id, const std::string& command, const std::string& message) -> std::expected<void, std::string>;
	
	// jobs
	auto consume_message_queue() -> std::expected<void, std::string>;
	auto check_global_message() -> std::expected<void, std::string>;
	auto clear_legacy_global_message_key() -> void;

	// message list
	auto request_client_status_update(const std::string& id, const std::string& sub_id, const std::string& message) -> std::expected<void, std::string>;

private:
	std::mutex mutex_;

	std::atomic_bool stopping_{ false };

	std::shared_ptr<NetworkServer> server_;
	std::shared_ptr<ThreadPool> thread_pool_;
	std::shared_ptr<Configurations> configurations_;

	// If you use more than 3 Redis clients, manage them.
	std::shared_ptr<RedisClient> redis_client_;

	std::string register_key_;

	std::map<std::string, std::function<std::expected<void, std::string>(const std::string&, const std::string&, const std::string&)>> messages_;
	const std::string global_message_key_ = "send_global_message";
};