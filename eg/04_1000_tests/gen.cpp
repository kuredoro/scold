#include <fstream>
#include <random>
#include <ctime>

using namespace std;

int main()
{
    ofstream file("inputs.txt");

    mt19937 rd(time(0));
    uniform_int_distribution<int64_t> dist(0, 1000100000);

    for (int i = 0; i < 1000; i++)
    {
        int64_t a = dist(rd), b = dist(rd);
        file << a << ' ' << b << "\n---\n";
        file << a + b << "\n===\n";
    }

    return 0;
}
