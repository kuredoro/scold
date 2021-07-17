#include <iostream>
#include <cmath>
#include <chrono>
#include <thread>

int main()
{
    int a;
    std::cin >> a;

    for (int i = 0; i < a * 500'000; i++)
        std::cos(50.0);

    std::cout << a << '\n';
    return 0;
}
