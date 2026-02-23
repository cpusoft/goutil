taskcycleutil: 完成抽象的周期状态机任务框架
1 每个任务有一个key，用于排重，有自定义数据结构，用于传输数据，有执行结果（ok或fail）
1.1 当任务从一个状态往下一个状态转换时，状态包括：待执行、执行中、执行完毕；
1.2 有一个禁止执行的map，map的key如果和任务的key相同，则禁止执行
1.3 框架每30分钟一周期
1.3.1 每周期的一开始：把所有状态为：待执行、执行成功和执行失败的任务，将其状态从待执行改为执行中，并开始异步执行，最长执行超时时间为2个30分钟+10分钟，即70分钟；
1.3.2 每周期过10分钟：检查一次执行结果，如果执行完毕，则执行后续任务。如果没有执行完，则下一个40分钟周期执行

2 如果任务在最长超时时间内执行完毕，状态都改为执行完毕
2.1 如果成功，则结果为ok，记录成功时间和结果，成功次数+1 ；
2.2 如果失败，则结果为fail，记录失败时间和原因，失败次数+1 
2.3 如果完成时，正好在某个周期的前10分钟内，则将在本周期执行后续任务。如果在某个周期后20分钟，则下个周期执行后续任务。

3 如果任务最长超时时间内内没有执行完毕
3.1 直接设为执行完毕，且结果为fail，记录失败时间和原因，失败次数+1


4 批量添加新任务，分两种情况，两种情况都需要支持，通过配置参数预先区分好，不会同时出现，
4.1 注意排重，不能与禁止执行map的key重复；不能与待任务、执行中任务、执行完毕的key重复。
4.2 仅仅从成功任务的结果中，解析得到新的任务，按4.1排重后，将这新任务，不等待下一个30分钟周期，状态设为执行中，立即异步执行，结果参照2和3处理。此而成功任务本身则改为待执行，需要下一个30分钟周期处理。  /*注释：即从tal取结果，递归处理*/
4.3 仅仅通过对外的接口，通过外部程序，时刻会增加新的任务，均为待处理执行，需要按4.1排重，均下一个30分钟周期处理  /*注释：即从precept注入任务，可能批量*/


# 运行所有测试
go test -v -timeout 120s

# 只运行单元测试
go test -v -run "Test.*" -timeout 60s

# 只运行临界值测试
go test -v -run "Test.*Critical|Test.*Limit|Test.*Zero|Test.*Nil|Test.*Invalid" -timeout 60s

# 只运行并发测试
go test -v -run "TestConcurrent.*" -timeout 60s

# 运行性能测试（含10w级别）
go test -bench=. -benchmem -timeout 300s

# 运行特定性能测试
go test -bench=BenchmarkAddTasks_100000 -benchmem -timeout 300s


# 运行所有长时间+大内存专项测试
go test -v -run "Test.*Long.*|Test.*Large.*|Test.*Memory.*" -timeout 300s

# 运行特定测试
go test -v -run TestLargeMemoryTask_Execution -timeout 120s

# 运行专项性能测试
go test -bench=Benchmark.*Long.*|Benchmark.*Large.* -benchmem -timeout 300s