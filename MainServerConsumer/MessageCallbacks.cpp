#include "MessageCallbacks.h"
#include "Logger.h"

#include <optional>


MessageCallbacks::MessageCallbacks()
{
	callbacks_.insert({"example_function", std::bind(&MessageCallbacks::example_callback_function, this, std::placeholders::_1)});
}

MessageCallbacks::~MessageCallbacks()
{

}

auto MessageCallbacks::message_call(const std::string& received_message) -> std::tuple<bool, std::optional<std::string>>
{
	return { true, std::nullopt };
}

auto MessageCallbacks::example_callback_function(const std::string& message) -> std::tuple<bool, std::optional<std::string>>
{
	return { true, std::nullopt };
}

