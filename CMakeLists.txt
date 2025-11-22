cmake_minimum_required(VERSION 3.18)

set(PROJECT_NAME RealTimeMessageChat)
set(CMAKE_CXX_STANDARD 20)
set(CMAKE_CXX_STANDARD_REQUIRED TRUE)
set(CMAKE_CXX_FLAGS_DEBUG "${CMAKE_CXX_FLAGS_DEBUG} -D_DEBUG")

set(CMAKE_POSITION_INDEPENDENT_CODE ON)

# Set the output directory to the build directory
set(CMAKE_RUNTIME_OUTPUT_DIRECTORY "${CMAKE_BINARY_DIR}/out")
set(CMAKE_LIBRARY_OUTPUT_DIRECTORY "${CMAKE_BINARY_DIR}/lib")
set(CMAKE_ARCHIVE_OUTPUT_DIRECTORY "${CMAKE_BINARY_DIR}/lib")

set(CMAKE_MODULE_PATH "${CMAKE_CURRENT_LIST_DIR}/cmake")

if(APPLE AND NOT CMAKE_CROSSCOMPILING)
    list(APPEND CMAKE_PREFIX_PATH /usr/local /opt/homebrew)
endif()

project(${PROJECT_NAME} VERSION 1.0.0.0)

find_package(Threads REQUIRED)
find_package(Boost REQUIRED COMPONENTS filesystem system json)

# Ensure Boost imported targets exist when toolchain/config generation differs
if(NOT TARGET Boost::filesystem)
    find_library(BOOST_FILESYSTEM_LIB NAMES boost_filesystem PATHS "${CMAKE_BINARY_DIR}/vcpkg_installed/${VCPKG_TARGET_TRIPLET}/lib" NO_DEFAULT_PATH)
    if(BOOST_FILESYSTEM_LIB)
        add_library(Boost::filesystem UNKNOWN IMPORTED)
        set_target_properties(Boost::filesystem PROPERTIES IMPORTED_LOCATION "${BOOST_FILESYSTEM_LIB}")
        target_include_directories(Boost::filesystem INTERFACE ${Boost_INCLUDE_DIRS})
    endif()
endif()
if(NOT TARGET Boost::system)
    find_library(BOOST_SYSTEM_LIB NAMES boost_system PATHS "${CMAKE_BINARY_DIR}/vcpkg_installed/${VCPKG_TARGET_TRIPLET}/lib" NO_DEFAULT_PATH)
    if(BOOST_SYSTEM_LIB)
        add_library(Boost::system UNKNOWN IMPORTED)
        set_target_properties(Boost::system PROPERTIES IMPORTED_LOCATION "${BOOST_SYSTEM_LIB}")
        target_include_directories(Boost::system INTERFACE ${Boost_INCLUDE_DIRS})
    endif()
endif()

set(BUILD_KAFKA_LIB OFF CACHE BOOL "Build kafka library")

add_subdirectory(.cpp_tool_kit)

add_subdirectory(CommonModule)
add_subdirectory(MainServer)
add_subdirectory(UserClient)
add_subdirectory(MainServerConsumer)
