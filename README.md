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
- 📝 **Detailed Logging**: Comprehensive logging of all operations for easier troubleshooting and auditing.
- 🔄 **Multi-Database Support**: Optimize multiple databases simultaneously with comparative analysis.
- 🌥️ **Cloud Storage**: Store optimization history in Amazon S3 for durability and scalability.

## 🎯 Use Cases

- 🚀 Optimize underperforming MongoDB deployments
- 🛡️ Proactively prevent performance degradation
- 💰 Reduce operational costs by improving resource utilization
- 🤖 Automate routine database maintenance tasks
- 💡 Gain insights into application query patterns and database usage

## 🏁 Getting Started

### Prerequisites

- Go 1.19 or later
- Access to a MongoDB database
- OpenAI API key for AI-driven optimizations
- (Optional) AWS credentials for S3 storage

### Installation

```bash
# Clone the repository
git clone https://github.com/theapemachine/lookatthatmongo.git
cd lookatthatmongo

# Build the application
go build -o lookatthatmongo

# Set your MongoDB connection string and OpenAI API key
export MONGO_URI="mongodb://username:password@hostname:port/database"
export OPENAI_API_KEY="your-openai-api-key"

# Run the optimizer
./lookatthatmongo --db yourDatabaseName
```

### Configuration

The application can be configured using environment variables or command-line flags:

#### Environment Variables

- `MONGO_URI`: MongoDB connection string (required)
- `MONGO_DB`: Database name to optimize (default: "FanAppDev2")
- `STORAGE_TYPE`: Storage type (file or s3) (default: "file")
- `STORAGE_PATH`: Path to store optimization history for file storage (default: "~/.lookatthatmongo/history")
- `LOG_LEVEL`: Logging level (debug, info, warn, error) (default: "info")
- `IMPROVEMENT_THRESHOLD`: Improvement threshold percentage (default: 5.0)
- `ENABLE_ROLLBACK`: Enable automatic rollback on failure (default: true)
- `MAX_OPTIMIZATIONS`: Maximum number of optimizations to apply (default: 3)
- `OPENAI_API_KEY`: Your OpenAI API key for AI-powered optimizations

#### S3 Storage Environment Variables

- `S3_BUCKET`: S3 bucket name (required when using S3 storage)
- `S3_REGION`: AWS region (default: "us-east-1")
- `S3_PREFIX`: Prefix for S3 objects (default: "optimization-records/")
- `S3_RETENTION_DAYS`: Number of days to retain records (default: 90)
- `S3_CREDENTIALS_FILE`: Path to AWS credentials file (optional)

#### Command-line Flags

- `--db`: MongoDB database name to optimize
- `--storage-type`: Storage type (file or s3)
- `--storage-path`: Path to store optimization history (for file storage)
- `--log-level`: Logging level (debug, info, warn, error)
- `--threshold`: Improvement threshold percentage
- `--enable-rollback`: Enable automatic rollback on failure
- `--max-optimizations`: Maximum number of optimizations to apply

#### S3 Storage Flags

- `--s3-bucket`: S3 bucket name
- `--s3-region`: AWS region
- `--s3-prefix`: Prefix for S3 objects
- `--s3-retention-days`: Number of days to retain records
- `--s3-credentials`: Path to AWS credentials file

Example:

```bash
export MONGO_URI="mongodb://username:password@hostname:port"
export OPENAI_API_KEY="your-openai-api-key"
./lookatthatmongo --db myDatabase --threshold 10.0 --log-level debug
```

### Using S3 Storage

To use S3 for storing optimization history:

```bash
export MONGO_URI="mongodb://username:password@hostname:port"
export OPENAI_API_KEY="your-openai-api-key"
export AWS_ACCESS_KEY_ID="your-access-key"
export AWS_SECRET_ACCESS_KEY="your-secret-key"

./lookatthatmongo --db myDatabase --storage-type s3 --s3-bucket my-optimization-bucket
```

### Multi-Database Optimization

To optimize multiple databases simultaneously:

```bash
export MONGO_URI="mongodb://username:password@hostname:port"
export OPENAI_API_KEY="your-openai-api-key"

./lookatthatmongo multi --databases db1,db2,db3 --parallel 3
```

Or using a comma-separated list:

```bash
./lookatthatmongo multi --db-list "db1,db2,db3" --parallel 3
```

### Cleanup Old Records

To cleanup old optimization records (particularly useful for S3 storage):

```bash
./lookatthatmongo cleanup --retention-days 30
```

## 🏗️ Architecture

Look At That Mon Go is built with a modular architecture that separates concerns and allows for easy extension:

### Core Components

- **MongoDB Connection**: Manages connections to MongoDB databases
- **Metrics Collection**: Gathers performance metrics from MongoDB
- **AI Analysis**: Uses AI to analyze metrics and suggest optimizations
- **Optimizer**: Applies optimizations to MongoDB
- **Tracker**: Tracks optimization history and measures impact
- **Storage**: Stores optimization history (both file-based and S3)
- **Config**: Manages application configuration from environment variables and flags
- **Logger**: Provides structured logging throughout the application

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

## 🧪 Testing

The project includes a comprehensive test suite to ensure reliability:

```bash
# Run all tests
go test ./...

# Run tests with coverage
go test -cover ./...

# Run tests for a specific package
go test ./storage
```

## 🚧 Project Status

This project is currently under active development. The core functionality is implemented and working, including:

- MongoDB connection and metrics collection
- AI-powered optimization suggestions
- Optimization application and validation
- Performance measurement and impact assessment
- File-based storage for optimization history
- Robust error handling and logging

Future enhancements will include a web dashboard, scheduled optimization runs, and support for multiple databases.

## 🤝 Contributing

Contributions are welcome! Please feel free to submit a Pull Request.

## 📝 License

This project is licensed under the MIT License - see the LICENSE file for details.
