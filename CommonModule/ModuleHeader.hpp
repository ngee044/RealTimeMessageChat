#pragma once

#include <vector>
#include <string>
#include <functional>
#include <optional>
#include <expected>

using client_combine_message_execute_callback = std::function<std::expected<void, std::string>(const std::string&, const std::string&, const std::string&, const std::vector<uint8_t>&)>;
using client_combine_message_parsing_callback = std::function<std::expected<void, std::string>(const std::string&, const std::string&, const std::string&, const std::string&, const std::vector<uint8_t>&)>;
using client_message_execute_callback = std::function<std::expected<void, std::string>(const std::string&, const std::string&, const std::string&)>;
using client_message_parsing_callback = std::function<std::expected<void, std::string>(const std::string&, const std::string&, const std::string&, const std::string&)>;

using server_combine_message_callback = std::function<std::expected<void, std::string>(const std::string&, const std::vector<uint8_t>&)>;
using server_combine_message_parsing_callback = std::function<std::expected<void, std::string>(const std::string&, const std::string&, const std::vector<uint8_t>&)>;
using server_message_execute_callback = std::function<std::expected<void, std::string>(const std::string&)>;
using server_message_parsing_callback = std::function<std::expected<void, std::string>(const std::string&, const std::string&)>;

