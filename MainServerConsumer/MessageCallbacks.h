#pragma once

#include <unordered_map>
#include <vector>
#include <string>
#include <tuple>
#include <optional>
#include <functional>


class MessageCallbacks final
{
public:
	MessageCallbacks();
	~MessageCallbacks();

	auto message_call(const std::string& received_message) -> std::tuple<bool, std::optional<std::string>>;

protected:
	auto example_callback_function(const std::string& message) -> std::tuple<bool, std::optional<std::string>>;

private:
	std::unordered_map<std::string, std::function<std::tuple<bool, std::optional<std::string>>(const std::string&)>> callbacks_;
};
