### 学习《用Go实现解释器和编译器》
---
### 语言特性
- 整型
- 布尔型
- 字符串
- 数组
- 哈希表
- 前缀、中缀、索引运算符
- 全局 | 局部变量绑定
- 表达式(1+1,1<1, 1!=1, 1==1...)
- return语句
- if语句
- 赋值
- 函数
- 高阶函数
- 内置函数 
- 简单宏实现

### 示例
- 变量绑定
```
let foo = "bar";
puts(foo); // bar

let a = 1;
a; // 1
```
- 数组
```
let arr = [1,2,3];
arr; // [1,2,3]
arr[0]; // 1 
```
- hash
```
let obj = {a:1};
obj; // {a:1};
obj["a"]; // 1
```
- if else
```
if(true){ 1 };
if(false){ 1 }else{ 2 };
if(1 < 2){ 1 }
if(1+1 > 1){ 2 }

let a = if(true){ 1 };
a; // 1
```
- function
```
let f = fn(x){ return 1;};
f(2); // 1

let a = 0;
a = 1;
puts(a); // 1
```


### TODO
- 赋值语句
```
let arr = [1,2,3];
arr[0] = 3;
arr[0]; // 3

let obj = {"a":1};
obj["a"] = 3;
obj["a"]; // 3
```
- 浮点数
```
let foo = 1.1;
puts(foo) //1.1
```
