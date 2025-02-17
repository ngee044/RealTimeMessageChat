#pragma once

#include <optional>
#include <map>
#include <mutex>
#include <memory>
#include <vector>
#include <string>

class UserClientManager
{
public:
	virtual ~UserClientManager(void);

private:
	UserClientManager(void);

#pragma region Handle
public:
	static UserClientManager& handle(void);

private:
	static std::unique_ptr<UserClientManager> handle_;
	static std::once_flag once_;
#pragma endregion
}