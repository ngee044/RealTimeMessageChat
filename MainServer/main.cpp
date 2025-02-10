#include "MainServer.h"
#include "Configurations.h"
#include "Logger.h"

#include "fmt/format.h"
#include "fmt/xchar.h"

#include <memory>
#include <string>
#include <signal.h>

std::shared_ptr<Configurations> configurations_ = nullptr;
std::shared_ptr<MainServer> main_server_ = nullptr;

auto main(int argc, char* argv[]) -> int
{
	configurations_ = std::make_shared<Configurations>(ArgumentParser(argc, argv));

	Logger::handle().file_mode(configurations_->write_file());
	Logger::handle().console_mode(configurations_->write_console());
	Logger::handle().write_interval(configurations_->write_interval());
	Logger::handle().log_root(configurations_->log_root_path());

	Logger::handle().start(configurations_->client_title());

	main_server_ = std::make_shared<MainServer>(configurations_);

	auto [success, message] = main_server_->start();
	if (!success)
	{
		Logger::handle().write(LogTypes::Error, message.value());
	}
	else
	{
		Logger::handle().write(LogTypes::Information, "MainServer started successfully");

		main_server_->wait_stop();
	}

	configurations_.reset();
	main_server_.reset();

	return 0;
}