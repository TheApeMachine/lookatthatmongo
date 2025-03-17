# 🚀 Look At That Mon Go

An AI-powered MongoDB optimization platform that intelligently analyzes, optimizes, and monitors your MongoDB deployments.

## 🔍 Overview

Look At That Mon Go leverages artificial intelligence to transform MongoDB performance tuning from an art into a science. The platform continuously monitors your MongoDB clusters, identifies performance bottlenecks, recommends and applies optimizations, and validates their impact—all with minimal human intervention.

## ✨ Features

- 📊 **Comprehensive Monitoring**: Collects detailed metrics on server performance, database statistics, collection usage, index efficiency, and query patterns.
- 🧠 **AI-Driven Analysis**: Uses advanced AI models to analyze performance data and identify optimization opportunities.
- 🔧 **Intelligent Optimization**: Automatically suggests and applies improvements to indexes, queries, schema design, and configuration settings.
- ✅ **Impact Validation**: Measures and reports the actual performance impact of each optimization.
- ↩️ **Automatic Rollback**: Safely reverts changes if they don't produce the expected improvements.
- 📈 **Historical Tracking**: Builds performance profiles over time to identify trends and recurring issues.

## 🎯 Use Cases

- 🚀 Optimize underperforming MongoDB deployments
- 🛡️ Proactively prevent performance degradation
- 💰 Reduce operational costs by improving resource utilization
- 🤖 Automate routine database maintenance tasks
- 💡 Gain insights into application query patterns and database usage

## 🏁 Getting Started

```bash
# Set your MongoDB connection string
export MONGO_URI="mongodb://username:password@hostname:port/database"

# Run the optimizer
go run main.go
```

### Configuration

You can configure the application using environment variables or command-line flags:

#### Environment Variables

- `MONGO_URI`: MongoDB connection string (required)
- `MONGO_DB`: Database name to optimize (default: "FanAppDev2")
- `STORAGE_PATH`: Path to store optimization history (default: "~/.lookatthatmongo/history")
- `LOG_LEVEL`: Logging level (debug, info, warn, error) (default: "info")
- `IMPROVEMENT_THRESHOLD`: Improvement threshold percentage (default: 5.0)
- `ENABLE_ROLLBACK`: Enable automatic rollback on failure (default: true)
- `MAX_OPTIMIZATIONS`: Maximum number of optimizations to apply (default: 3)

#### Command-line Flags

- `--db`: MongoDB database name to optimize
- `--storage`: Path to store optimization history
- `--log-level`: Logging level (debug, info, warn, error)
- `--threshold`: Improvement threshold percentage
- `--enable-rollback`: Enable automatic rollback on failure
- `--max-optimizations`: Maximum number of optimizations to apply

Example:

```bash
export MONGO_URI="mongodb://username:password@hostname:port"
go run main.go --db myDatabase --threshold 10.0 --log-level debug
```

## 🏗️ Architecture

Look At That Mon Go is built with a modular architecture that separates concerns and allows for easy extension:

### Core Components

- **MongoDB Connection**: Manages connections to MongoDB databases
- **Metrics Collection**: Gathers performance metrics from MongoDB
- **AI Analysis**: Uses AI to analyze metrics and suggest optimizations
- **Optimizer**: Applies optimizations to MongoDB
- **Tracker**: Tracks optimization history and measures impact
- **Storage**: Stores optimization history for future reference

### Workflow

1. **Collect Metrics**: Gather performance metrics from MongoDB
2. **Generate Suggestions**: Use AI to analyze metrics and suggest optimizations
3. **Apply Optimizations**: Apply the suggested optimizations to MongoDB
4. **Measure Impact**: Collect metrics again and measure the impact
5. **Take Action**: Based on the impact, take appropriate action (continue, alert, rollback)
6. **Store History**: Store the optimization history for future reference

### Error Handling

The application includes robust error handling to ensure database safety:

- **Validation**: Validates optimization suggestions before applying
- **Rollback**: Automatically rolls back changes if they don't improve performance
- **Logging**: Comprehensive logging of all operations and errors
- **Recovery**: Graceful recovery from errors during optimization

## 🚧 Project Status

This project is currently under active development. Contributions and feedback are welcome!

## 📝 License

This project is licensed under the MIT License - see the LICENSE file for details.
