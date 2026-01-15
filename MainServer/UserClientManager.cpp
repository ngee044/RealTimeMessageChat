#include "UserClientManager.h"

#include "Logger.h"

#include <format>


using namespace Utilities;

UserClientManager::UserClientManager()
{

}

UserClientManager::~UserClientManager()
{
}

auto UserClientManager::add(const std::string& id, const std::string& sub_id) -> std::tuple<bool, std::optional<std::string>>
{
	std::scoped_lock lock(mutex_);

	auto iter = clients_.find({ id, sub_id });
	if (iter != clients_.end())
	{
		Logger::handle().write(LogTypes::Error, std::format("Client is already exist: {}, {}", id, sub_id));
		return { false, std::format("Client is already exist: {}, {}", id, sub_id) };
	}

	clients_.insert({ {id, sub_id}, {"", ""} });

	return { true, std::nullopt };
}

auto UserClientManager::remove(const std::string& id, const std::string& sub_id) -> std::tuple<bool, std::optional<std::string>>
{
	std::scoped_lock lock(mutex_);

	auto iter = clients_.find({ id, sub_id });
	if (iter == clients_.end())
	{
		Logger::handle().write(LogTypes::Error, std::format("Client is not exist: {}, {}", id, sub_id));
		return { false, std::format("Client is not exist: {}, {}", id, sub_id) };
	}

	clients_.erase(iter);

	return { true, std::nullopt };
}

auto UserClientManager::clinets() -> std::map<std::tuple<std::string, std::string>, std::tuple<std::string, std::string>>
{
	std::scoped_lock lock(mutex_);
	return clients_;  // Return a copy for thread safety
}

#pragma region Handle
std::unique_ptr<UserClientManager> UserClientManager::handle_ = nullptr;
std::once_flag UserClientManager::once_;

auto UserClientManager::handle(void) -> UserClientManager&
{
	std::call_once(once_, []() {
		handle_.reset(new UserClientManager);
		});

	return *handle_;
}
#pragma endregion