#include "MainServer.h"

MainServer::MainServer(std::shared_ptr<Configurations> configurations)
{
}

MainServer::~MainServer(void)
{
}

auto MainServer::start() -> std::tuple<bool, std::optional<std::string>>
{
	return std::tuple<bool, std::optional<std::string>>();
}

auto MainServer::stop() -> std::tuple<bool, std::optional<std::string>>
{
	return std::tuple<bool, std::optional<std::string>>();
}

auto MainServer::wait_stop() -> std::tuple<bool, std::optional<std::string>>
{
	return std::tuple<bool, std::optional<std::string>>();
}

auto MainServer::create_thread_pool() -> std::tuple<bool, std::optional<std::string>>
{
	return std::tuple<bool, std::optional<std::string>>();
}

auto MainServer::destroy_thread_pool() -> void
{
}
