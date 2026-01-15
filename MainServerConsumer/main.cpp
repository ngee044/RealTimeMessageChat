#include "Configurations.h"
#include "MainServerConsumer.h"
#include "Logger.h"

#include <format>

#include <memory>
#include <string>
#include <signal.h>

using namespace Utilities;

void register_signal(void);
void deregister_signal(void);
void signal_callback(int32_t signum);

std::shared_ptr<Configurations> configurations_ = nullptr;
std::shared_ptr<MainServerConsumer> main_server_consumer = nullptr;

auto main(int argc, char* argv[]) -> int
{
	register_signal();

	configurations_ = std::make_shared<Configurations>(ArgumentParser(argc, argv));

	Logger::handle().file_mode(configurations_->write_file());
	Logger::handle().console_mode(configurations_->write_console());
	Logger::handle().write_interval(configurations_->write_interval());
	Logger::handle().log_root(configurations_->log_root_path());

	Logger::handle().start(configurations_->client_title());

	main_server_consumer = std::make_shared<MainServerConsumer>(configurations_);

	auto [success, message] = main_server_consumer->start();
	if (!success)
	{
		Logger::handle().write(LogTypes::Error, message.value());
	}
	else
	{
		Logger::handle().write(LogTypes::Information, "MainServerConsumer started successfully");
	}
	main_server_consumer->wait_stop();
	main_server_consumer.reset();

	configurations_.reset();

	Logger::handle().stop();
	Logger::destroy();

	deregister_signal();

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

	if (main_server_consumer == nullptr)
	{
		return;
	}

	Logger::handle().write(LogTypes::Information, std::format("attempt to stop MainServerConsumer from signal {}", signum));
	main_server_consumer->stop();
}