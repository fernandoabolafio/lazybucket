# LazyBucket Usage Guide

## Prerequisites

Before using LazyBucket, make sure you have:

1. Google Cloud SDK installed
2. Authenticated with Google Cloud
3. A Google Cloud Project ID

```bash
# Install Google Cloud SDK (if not already installed)
# Follow instructions at: https://cloud.google.com/sdk/docs/install

# Authenticate with Google Cloud
gcloud auth application-default login

# Set your Google Cloud Project ID (replace 'your-project-id' with your actual project ID)
export GOOGLE_CLOUD_PROJECT=your-project-id
```

## Running LazyBucket

You can run LazyBucket in two ways:

1. Using the environment variable (recommended):

```bash
export GOOGLE_CLOUD_PROJECT=your-project-id
./lazybucket
```

2. Using the command line flag:

```bash
./lazybucket --project=your-project-id
```

## Keyboard Shortcuts

| Key           | Action                      |
| ------------- | --------------------------- |
| ↑ / k         | Move cursor up              |
| ↓ / j         | Move cursor down            |
| Enter         | Open selected bucket/folder |
| Backspace / b | Go back to parent directory |
| v             | View file content           |
| r             | Refresh current view        |
| ? / h         | Toggle help view            |
| q / Ctrl+C    | Quit application            |

## Navigation

1. **Main View**: When you start the application, you'll see a list of all your GCS buckets.
2. **Bucket Navigation**: Select a bucket and press Enter to view its contents.
3. **Folder Navigation**: Navigate through folders by selecting them and pressing Enter.
4. **Going Back**: Press Backspace or 'b' to go back to the parent directory.
5. **Viewing Files**: Select a file and press 'v' to view its contents.

## Tips

- Use 'r' to refresh the current view if you've made changes to your buckets outside the application.
- The path at the top of the screen shows your current location in the bucket hierarchy.
- The status bar at the bottom shows information about the current operation.

## Troubleshooting

If you encounter authentication issues:

1. Make sure you're authenticated with Google Cloud:

   ```bash
   gcloud auth application-default login
   ```

2. Verify you have the necessary permissions to access the buckets.

3. Check your internet connection.

If the application crashes or behaves unexpectedly, please report the issue on GitHub with details about what you were doing when the problem occurred.
