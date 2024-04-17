#include <fstream>
#include <string>
#include <vector>

int main(int argc, char **argv)
{
    std::string path = "";

    int test_count = std::stoi(argv[1]);
    int ticker_count = std::stoi(argv[2]);
    int command_count = std::stoi(argv[3]);

    for (int i = 0; i < test_count; i++)
    {
        // create file
        std::string test_name = std::to_string(i) + ".in";
        std::ofstream test_file;
        test_file.open(path + test_name);

        int order_id = 1;
        std::vector<std::vector<int>> threadToIds(40);

        test_file << 40 << std::endl;
        test_file << "o" << std::endl;
        for (int j = 0; j < command_count; j++)
        {
            int thread = rand() % 40;
            int buy_or_sell_or_canc = rand() % 3;
            int ticker = rand() % ticker_count;

            if (buy_or_sell_or_canc == 2)
            {
                if (threadToIds[thread].empty())
                    continue;
                test_file << thread << " C " << threadToIds[thread][rand() % threadToIds[thread].size()] << std::endl;
                continue;
            }

            int id = order_id++;
            threadToIds[thread].push_back(id);
            test_file << thread << (buy_or_sell_or_canc == 0 ? " B " : " S ") << id << " " << ticker << " " << (rand() % 100 + 1) << " "
                      << (rand() % 100 + 1) << std::endl;
        }
        test_file << "x" << std::endl;
        test_file.close();
    }
}