#include "UserClient.h"

UserClient::UserClient(std::shared_ptr<Configurations> configurations)
{
}

UserClient::~UserClient(void)
{
}

auto UserClient::start() -> std::tuple<bool, std::optional<std::string>>
{
	return std::tuple<bool, std::optional<std::string>>();
}

auto UserClient::stop() -> std::tuple<bool, std::optional<std::string>>
{
	return std::tuple<bool, std::optional<std::string>>();
}

auto UserClient::wait_stop() -> std::tuple<bool, std::optional<std::string>>
{
	return std::tuple<bool, std::optional<std::string>>();
}
