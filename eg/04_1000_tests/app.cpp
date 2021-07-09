#include <iostream>
#include <cmath>

int main()
{
    int a, b;
    std::cin >> a >> b;

    double x = 50;
    for (int i = 0; i < 10000000; i++)
    {
        x = std::cos(x);
    }

    std::cout << a + b << '\n';
    return 0;
}
