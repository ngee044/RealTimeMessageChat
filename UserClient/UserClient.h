#pragma once

#include "Configurations.h"

#include "NetworkClient.h"

#include <string>
#include <memory>
#include <tuple>
#include <optional>

class UserClient
{
public:
	UserClient(std::shared_ptr<Configurations> configurations);
	~UserClient(void);

	auto start() -> std::tuple<bool, std::optional<std::string>>;
	auto stop() -> std::tuple<bool, std::optional<std::string>>;
	auto wait_stop() -> std::tuple<bool, std::optional<std::string>>;	

protected:

private:

};