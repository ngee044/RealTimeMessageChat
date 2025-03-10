#include "MainServer.h"
#include "Configurations.h"
#include "Logger.h"

#include "fmt/format.h"
#include "fmt/xchar.h"

#include <memory>
#include <string>
#include <signal.h>

using namespace Utilities;

void register_signal(void);
void deregister_signal(void);
void signal_callback(int32_t signum);

std::shared_ptr<Configurations> configurations_ = nullptr;
std::shared_ptr<MainServer> server_ = nullptr;

auto main(int argc, char* argv[]) -> int
{
	configurations_ = std::make_shared<Configurations>(ArgumentParser(argc, argv));

	Logger::handle().file_mode(configurations_->write_file());
	Logger::handle().console_mode(configurations_->write_console());
	Logger::handle().write_interval(configurations_->write_interval());
	Logger::handle().log_root(configurations_->log_root_path());

	Logger::handle().start(configurations_->client_title());

	server_ = std::make_shared<MainServer>(configurations_);

	auto [success, message] = server_->start();
	if (!success)
	{
		Logger::handle().write(LogTypes::Error, message.value());
	}
	else
	{
		Logger::handle().write(LogTypes::Information, "MainServer started successfully");

		server_->wait_stop();
	}
	server_.reset();
	configurations_.reset();

	Logger::handle().stop();
	Logger::destroy();

	return 0;
}

void register_signal(void)
{
	signal(SIGINT, signal_callback);
	signal(SIGILL, signal_callback);
	signal(SIGABRT, signal_callback);
	signal(SIGFPE, signal_callback);
	signal(SIGSEGV, signal_callback);
	signal(SIGTERM, signal_callback);
}

void deregister_signal(void)
{
	signal(SIGINT, nullptr);
	signal(SIGILL, nullptr);
	signal(SIGABRT, nullptr);
	signal(SIGFPE, nullptr);
	signal(SIGSEGV, nullptr);
	signal(SIGTERM, nullptr);
}

void signal_callback(int32_t signum)
{
	deregister_signal();

	if (server_ == nullptr)
	{
		return;
	}

	Logger::handle().write(LogTypes::Information, fmt::format("attempt to stop AudioCalculator from signal {}", signum));
	server_->stop();
}