rm -rf build

mkdir build
cd build
cmake .. -DCAMKE_TOOLCHAIN_FILE="../../vcpkg/scripts/buildsystems/vcpkg.cmake" -DCAMKE_BUILD_TYPE=Release -DBUILD_SHARED_LIBS=OFF