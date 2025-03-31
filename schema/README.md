# FlatBuffers Schema 说明文档

## 简介

本目录包含了DESKTK项目客户端（UE）和服务器（GO）之间通信所使用的FlatBuffers Schema定义文件。这些.fbs文件定义了双方通信的数据结构，通过FlatBuffers编译器生成相应的C++和Go代码。

## 文件说明

- `game.fbs` - 游戏基础数据结构定义
- `C2SCommand.fbs` - 客户端发送到服务器的命令结构
- `S2CCommand.fbs` - 服务器发送到客户端的命令结构

## 使用方法

### 编译Schema生成UE客户端代码

我们提供了`CompileFlatbufferForUE.bat`脚本来自动编译Schema文件并生成C++代码。该脚本会将生成的代码放置到UE插件的正确目录中。

#### 初次使用

1. 双击运行`CompileFlatbufferForUE.bat`
2. 脚本会创建一个配置文件`fbs_config.txt`，并自动用记事本打开
3. 在配置文件中设置UE项目的根路径，默认为`C:\UEProjects\DESKTK`
4. 保存配置文件后再次运行脚本

#### 常规使用

1. 修改Schema文件（.fbs）后，运行`CompileFlatbufferForUE.bat`
2. 脚本会自动清理旧文件，编译所有Schema文件，并生成相应的C++代码
3. 生成的文件会被放置在`UE项目路径\Plugins\UFlatBuffers\Source\UFlatBuffers\Generated\`目录下
4. 脚本还会生成一个`FBSGenerated.h`主头文件，包含所有生成的头文件

### 为GO服务器生成代码

GO服务器端代码需要单独生成，可以使用以下命令：

```bash
flatc --go -o path/to/server/generated schema/*.fbs
```

## 注意事项

1. 确保已安装FlatBuffers编译器（flatc），并已添加到系统PATH中
2. 修改Schema文件后需要重新生成代码
3. 需确保客户端和服务器使用完全相同的Schema文件
4. 修改Schema时注意向后兼容性，避免破坏现有通信

## FlatBuffers编译选项说明

脚本使用了以下FlatBuffers编译选项：

- `--cpp` - 生成C++代码
- `--cpp-std c++17` - 使用C++17标准
- `--scoped-enums` - 使用作用域枚举（C++11的enum class）
- `--gen-object-api` - 生成对象API，便于创建和操作FlatBuffers对象
- `--gen-compare` - 生成比较函数
- `--cpp-ptr-type UniquePtr` - 使用std::unique_ptr作为指针类型
- `--cpp-static-reflection` - 生成静态反射代码
- `--filename-suffix ""` - 不添加文件名后缀

## 问题排查

### 常见问题

1. **找不到flatc编译器**
   - 确保已安装FlatBuffers并添加到PATH环境变量中

2. **无法找到UE项目路径**
   - 检查`fbs_config.txt`中的UE_PROJECT_DIR设置是否正确

3. **无法创建输出目录**
   - 确保UE项目路径下存在UFlatBuffers插件
   - 检查当前用户是否有写入权限

4. **编译错误**
   - 检查Schema文件语法是否正确
   - 查看错误信息以获取详细信息 