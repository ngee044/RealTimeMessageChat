#include "Configurations.h"

#include "File.h"
#include "Logger.h"
#include "Converter.h"

#include "fmt/xchar.h"
#include "fmt/format.h"

#include "boost/json.hpp"
#include "boost/json/parse.hpp"

#include <filesystem>


Configurations::Configurations(ArgumentParser&& arguments)
	: write_file_(LogTypes::None)
	, write_console_(LogTypes::Information)
	, console_windows_(false)
	, callback_message_log_(LogTypes::Error)
	, root_path_("")
	, high_priority_count_(3)
	, normal_priority_count_(3)
	, low_priority_count_(5)
	, write_interval_(1000)
	, log_root_path_("")
	, buffer_size_(1024)
	, server_ip_("127.0.0.1")
	, server_port_(9876)
	, encrypt_mode_(true)
{
	root_path_ = arguments.program_folder();

	load();
	parse(arguments);
}

Configurations::~Configurations(void) {}

auto Configurations::write_file() -> LogTypes { return write_file_; }

auto Configurations::encrypt_mode() -> bool { return false; }

auto Configurations::write_console() -> LogTypes { return write_console_; }

auto Configurations::console_windows() -> bool { return console_windows_; }

auto Configurations::high_priority_count() -> uint16_t { return high_priority_count_; }

auto Configurations::normal_priority_count() -> uint16_t { return normal_priority_count_; }

auto Configurations::low_priority_count() -> uint16_t { return low_priority_count_; }

auto Configurations::write_interval() -> uint16_t { return write_interval_; }

auto Configurations::client_title() -> std::string { return client_title_; }

auto Configurations::log_root_path() -> std::string { return log_root_path_; }

auto Configurations::buffer_size() -> std::size_t { return buffer_size_; }

auto Configurations::server_ip() -> std::string { return server_ip_; }

auto Configurations::server_port() -> uint16_t { return server_port_; }

auto Configurations::load() -> void
{
	std::filesystem::path path = root_path_ + "user_client_configurations.json";
	if (!std::filesystem::exists(path))
	{
		Logger::handle().write(LogTypes::Error, fmt::format("Configurations file does not exist: {}", path.string()));
		return;
	}

	File source;
	source.open(fmt::format("{}user_client_configurations.json", root_path_), std::ios::in | std::ios::binary, std::locale(""));
	auto [source_data, error_message] = source.read_bytes();
	if (source_data == std::nullopt)
	{
		Logger::handle().write(LogTypes::Error, error_message.value());
		return;
	}

	boost::json::object message = boost::json::parse(Converter::to_string(source_data.value())).as_object();

	if (message.contains("client_title") && message.at("client_title").is_string())
	{
		client_title_ = message.at("client_title").as_string().data();
	}

	if (message.contains("log_root_path") && message.at("log_root_path").is_string())
	{
		log_root_path_ = message.at("log_root_path").as_string().data();
	}

	if (message.contains("write_file") && message.at("write_file").is_string())
	{
		write_file_ = static_cast<LogTypes>(message.at("write_file_log").as_int64());
	}

	if (message.contains("write_console") && message.at("write_console").is_string())
	{
		write_console_ = static_cast<LogTypes>(message.at("write_console").as_int64());
	}

	if (message.contains("callback_message_log") && message.at("callback_message_log").is_string())
	{
		callback_message_log_ = static_cast<LogTypes>(message.at("callback_message_log").as_int64());
	}

	if (message.contains("console_windows") && message.at("console_windows").is_bool())
	{
		console_windows_ = message.at("console_windows").as_bool();
	}

	if (message.contains("high_priority_count") && message.at("high_priority_count").is_number())
	{
		high_priority_count_ = static_cast<int>(message.at("high_priority_count").as_int64());
	}

	if (message.contains("normal_priority_count") && message.at("normal_priority_count").is_number())
	{
		normal_priority_count_ = static_cast<int>(message.at("normal_priority_count").as_int64());
	}

	if (message.contains("low_priority_count") && message.at("low_priority_count").is_number())
	{
		low_priority_count_ = static_cast<int>(message.at("low_priority_count").as_int64());
	}

	if (message.contains("write_interval") && message.at("write_interval").is_number())
	{
		write_interval_ = static_cast<int>(message.at("write_interval").as_int64());
	}

	if (message.contains("buffer_size") && message.at("buffer_size").is_number())
	{
		buffer_size_ = static_cast<int>(message.at("buffer_size").as_int64());
	}

	if (message.contains("main_server_ip") && message.at("main_server_ip").is_string())
	{
		server_ip_ = message.at("main_server_ip").as_string().data();
	}

	if (message.contains("main_server_port") && message.at("main_server_port").is_number())
	{
		server_port_ = static_cast<int>(message.at("main_server_port").as_int64());
	}
}

auto Configurations::parse(ArgumentParser& arguments) -> void
{
	auto string_target = arguments.to_string("--client_title");
	if (string_target != std::nullopt)
	{
		client_title_ = string_target.value();
	}

	string_target = arguments.to_string("--log_root_path");
	if (string_target != std::nullopt)
	{
		log_root_path_ = string_target.value();
	}

	auto ushort_target = arguments.to_ushort("--write_interval");
	if (ushort_target != std::nullopt)
	{
		write_interval_ = ushort_target.value();
	}

	auto int_target = arguments.to_int("--write_console_log");
	if (int_target != std::nullopt)
	{
		write_console_ = (LogTypes)int_target.value();
	}

	int_target = arguments.to_int("--write_file_log");
	if (int_target != std::nullopt)
	{
		write_file_ = (LogTypes)int_target.value();
	}
}