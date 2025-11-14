#include <iostream>
#include <vector>
#include <ranges>

int main() {
    // C++20 ranges example
    std::vector<int> numbers = {1, 2, 3, 4, 5};

    auto even_numbers = numbers
        | std::views::filter([](int n) { return n % 2 == 0; });

    std::cout << "Even numbers: ";
    for (int n : even_numbers) {
        std::cout << n << " ";
    }
    std::cout << std::endl;

    return 0;
}
