#pragma once

#include "Configurations.h"

#include "NetworkServer.h"

#include <string>
#include <memory>
#include <tuple>
#include <optional>

using namespace Thread;
using namespace Network;

class MainServer
{
public:
	MainServer(std::shared_ptr<Configurations> configurations);
	~MainServer(void);

	auto start() -> std::tuple<bool, std::optional<std::string>>;
	auto stop() -> std::tuple<bool, std::optional<std::string>>;
	auto wait_stop() -> std::tuple<bool, std::optional<std::string>>;

protected:
	auto create_thread_pool() -> std::tuple<bool, std::optional<std::string>>;
	auto destroy_thread_pool() -> void;

private:
	std::mutex mutex_;

	std::shared_ptr<NetworkServer> server_;
	std::shared_ptr<ThreadPool> thread_pool_;
	std::shared_ptr<Configurations> configurations_;

	std::map<std::string, std::function<std::tuple<bool, std::optional<std::string>>(const std::string&)>> messages_;


};