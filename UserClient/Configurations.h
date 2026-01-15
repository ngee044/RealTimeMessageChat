#pragma once

#include "LogTypes.h"
#include "ArgumentParser.h"

#include <string>
#include <map>
#include <optional>

using namespace Utilities;

class Configurations
{
public:
	Configurations(ArgumentParser&& arguments);
	virtual ~Configurations(void);

	auto write_file() -> LogTypes;
	auto encrypt_mode() -> bool;
	auto write_console() -> LogTypes;
	auto console_windows() -> bool;

	auto high_priority_count() -> uint16_t;
	auto normal_priority_count() -> uint16_t;
	auto low_priority_count() -> uint16_t;
	auto write_interval() -> uint16_t;

	auto client_title() -> std::string;
	auto log_root_path() -> std::string;

	auto buffer_size() -> std::size_t;
	auto server_ip() -> std::string;
	auto server_port() -> uint16_t;


protected:
	auto load() -> void;
	auto parse(ArgumentParser& arguments) -> void;
	auto validate_configuration() -> void;

private:
	LogTypes write_file_;
	LogTypes write_console_;
	LogTypes callback_message_log_;
	bool console_windows_;
	bool encrypt_mode_;

	uint16_t high_priority_count_;
	uint16_t normal_priority_count_;
	uint16_t low_priority_count_;
	uint16_t write_interval_;

	std::string client_title_;
	std::string root_path_;
	std::string log_root_path_;

	std::size_t buffer_size_;
	std::string server_ip_;
	uint16_t server_port_;
};