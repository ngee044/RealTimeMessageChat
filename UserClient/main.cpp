#include "UserClient.h"
#include "Configurations.h"
#include "Logger.h"

#include "fmt/format.h"
#include "fmt/xchar.h"

#include <memory>
#include <string>
#include <signal.h>

std::shared_ptr<Configurations> configurations_ = nullptr;
std::shared_ptr<UserClient> user_client_ = nullptr;

auto main(int argc, char* argv[]) -> int
{
	configurations_ = std::make_shared<Configurations>(ArgumentParser(argc, argv));

	Logger::handle().file_mode(configurations_->write_file());
	Logger::handle().console_mode(configurations_->write_console());
	Logger::handle().write_interval(configurations_->write_interval());
	Logger::handle().log_root(configurations_->log_root_path());

	Logger::handle().start(configurations_->client_title());

	user_client_ = std::make_shared<UserClient>(configurations_);

	auto [success, message] = user_client_->start();
	if (!success)
	{
		Logger::handle().write(LogTypes::Error, message.value());
	}
	else
	{
		Logger::handle().write(LogTypes::Information, "UserClient started successfully");

		user_client_->wait_stop();
	}

	configurations_.reset();
	user_client_.reset();

	return 0;
}