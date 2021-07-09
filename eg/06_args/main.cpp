#include <iostream>

using namespace std;

int main(int argc, char* argv[])
{
    if (argc != 2)
        return -1;

    int surplus = stoi(argv[1]);

    int a;
    cin >> a;
    cout << a + surplus << '\n';

    return 0;
}
