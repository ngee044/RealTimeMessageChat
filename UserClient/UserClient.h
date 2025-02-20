#pragma once

#include "Configurations.h"

#include "NetworkClient.h"
#include "ThreadPool.h"

#include <string>
#include <memory>
#include <tuple>
#include <optional>
#include <functional>
#include <map>

using namespace Thread;
using namespace Network;

class UserClient
{
public:
	UserClient(std::shared_ptr<Configurations> configurations);
	~UserClient(void);

	auto start() -> std::tuple<bool, std::optional<std::string>>;
	auto stop() -> void;
	
protected:
	auto create_thread_pool() -> std::tuple<bool, std::optional<std::string>>;
	auto destroy_thread_pool() -> void;

	auto received_connection(const bool& condition, const bool& by_itself) -> std::tuple<bool, std::optional<std::string>>;
	auto received_message(const std::string& message) -> std::tuple<bool, std::optional<std::string>>;

	auto parsing_message(const std::string& command, const std::string& message) -> std::tuple<bool, std::optional<std::string>>;
	auto update_user_clinet_status(const std::string message) -> std::tuple<bool, std::optional<std::string>>;

private:
	std::mutex mutex_;

	std::shared_ptr<NetworkClient> client_;
	std::shared_ptr<ThreadPool> thread_pool_;
	std::shared_ptr<Configurations> configurations_;

	std::string register_key_;

	std::map<std::string, std::function<std::tuple<bool, std::optional<std::string>>(const std::string&)>> messages_;

};