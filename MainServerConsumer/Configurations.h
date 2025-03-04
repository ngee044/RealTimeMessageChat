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

	auto redis_host() -> std::string;
	auto redis_port() -> int;
	auto redis_ttl_sec() -> int;
	auto redis_db_user_status_index() -> int;
	auto redis_db_global_message_index() -> int;
	auto use_redis() -> bool;
	auto use_redis_tls() -> bool;
	auto global_message_key() -> std::string;

	auto rabbit_mq_host() -> std::string;
	auto rabbit_mq_port() -> int;
	auto rabbit_mq_user_name() -> std::string;
	auto rabbit_mq_password() -> std::string;

	auto use_ssl() -> bool;
	auto ca_cert() -> std::string;
	auto engine() -> std::string;
	auto client_cert() -> std::string;
	auto client_key() -> std::string;

	auto consume_queue_name() -> std::string;

protected:
	auto load() -> void;
	auto parse(ArgumentParser& arguments) -> void;

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

	std::string redis_host_;
	int redis_port_;
	int redis_ttl_sec_;
	int redis_db_global_message_index_;
	int redis_db_user_status_index_;
	bool use_redis_;
	bool use_redis_tls_;

	std::string global_message_key_;

	std::string consume_queue_name_;
	std::string rabbit_mq_host_;
	int rabbit_mq_port_;
	std::string rabbit_mq_user_name_;
	std::string rabbit_mq_password_;

	bool use_ssl_;
	std::string ca_cert_;
	std::string engine_;
	std::string client_cert_;
	std::string client_key_;
};