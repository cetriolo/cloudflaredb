# Web Assets

This directory contains static web assets served by the API application.

## Contents

- `static/index.html` - Interactive API testing page

## API Testing Page

The `index.html` file provides a browser-based interface for testing all API endpoints.

### Features

- Visual interface with color-coded HTTP methods
- Real-time request/response testing
- JSON response formatting
- Form validation
- Loading indicators
- Status badges

### Access

Once the server is running, the testing page is available at:

```
http://localhost:8080
```

### Technologies

- Pure HTML/CSS/JavaScript
- No external dependencies
- No build process required
- Works in all modern browsers

### Customization

You can customize the testing page by editing `static/index.html`. The file contains:

- HTML structure in the `<body>` section
- CSS styles in the `<style>` section
- JavaScript logic in the `<script>` section

### Security Note

⚠️ This testing page is intended for development use only. For production deployments, consider:

1. Removing the static file server
2. Adding authentication/authorization
3. Restricting access by environment
4. Using a proper API documentation tool like Swagger

## Adding More Pages

To add additional static pages:

1. Create new HTML files in the `static/` directory
2. Access them at `http://localhost:8080/filename.html`
3. The Go file server automatically serves all files in this directory

## Documentation

For detailed usage instructions, see [docs/API_TESTER.md](../docs/API_TESTER.md).
