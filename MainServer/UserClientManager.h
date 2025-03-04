#pragma once

#include <map>
#include <mutex>
#include <string>
#include <tuple>
#include <optional>


class UserClientManager
{
public:
	virtual ~UserClientManager();

	auto add(const std::string& id, const std::string& sub_id) -> std::tuple<bool, std::optional<std::string>>;
	auto remove(const std::string& id, const std::string& sub_id) -> std::tuple<bool, std::optional<std::string>>;

	auto clinets() -> const std::map<std::tuple<std::string, std::string>, std::tuple<std::string, std::string>>&;
private:
	UserClientManager();

	std::mutex mutex_;
	std::map<std::tuple<std::string, std::string>, std::tuple<std::string, std::string>> clients_;

#pragma region Handle
public:
	static UserClientManager& handle();

private:
	static std::unique_ptr<UserClientManager> handle_;
	static std::once_flag once_;

#pragma endregion
};