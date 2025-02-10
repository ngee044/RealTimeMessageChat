#pragma once

#include "Configurations.h"

#include "NetworkServer.h"

#include <string>
#include <memory>
#include <tuple>
#include <optional>

class MainServer
{
public:
	MainServer(std::shared_ptr<Configurations> configurations);
	~MainServer(void);

	auto start() -> std::tuple<bool, std::optional<std::string>>;
	auto stop() -> std::tuple<bool, std::optional<std::string>>;
	auto wait_stop() -> std::tuple<bool, std::optional<std::string>>;

protected:

private:

};