#include "UserClientManager.h"

#include "Logger.h"
#include "Converter.h"
#include "LogTypes.h"

#include "fmt/format.h"
#include "fmt/xchar.h"

#include "boost/json.hpp"
#include "boost/json/parse.hpp"

#include <filesystem>

UserClientManager::UserClientManager()
{

}

UserClientManager::~UserClientManager()
{
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