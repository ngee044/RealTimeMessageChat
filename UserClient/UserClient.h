#pragma once

#include "Configurations.h"

#include "NetworkClient.h"
#include "ThreadPool.h"

#include <string>
#include <memory>
#include <atomic>
#include <tuple>
#include <optional>
#include <expected>
#include <functional>
#include <map>

using namespace Thread;
using namespace Network;

class UserClient
{
public:
	UserClient(std::shared_ptr<Configurations> configurations);
	~UserClient(void);

	auto start() -> std::expected<void, std::string>;
	auto stop() -> void;
	
protected:
	auto create_thread_pool() -> std::expected<void, std::string>;
	auto destroy_thread_pool() -> void;

	auto received_connection(bool condition, bool by_itself) -> std::expected<void, std::string>;
	auto received_message(const std::string& message) -> std::expected<void, std::string>;

	auto parsing_message(const std::string& command, const std::string& message) -> std::expected<void, std::string>;
	auto update_user_clinet_status(const std::string message) -> std::expected<void, std::string>;
	auto send_broadcast_message(const std::string message) -> std::expected<void, std::string>;

private:
	std::mutex mutex_;

	std::atomic_bool stopping_{ false };

	std::shared_ptr<NetworkClient> client_;
	std::shared_ptr<ThreadPool> thread_pool_;
	std::shared_ptr<Configurations> configurations_;

	std::string register_key_;

	std::map<std::string, std::function<std::expected<void, std::string>(const std::string&)>> messages_;

};