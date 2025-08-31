# CLAUDE.md - UI Directory

This file provides guidance to Claude Code when working with the Flipt UI application.

## UI Overview

The Flipt UI is a React-based single-page application that provides a web interface for managing feature flags. It's built with modern technologies and gets embedded into the main Flipt server binary.

## Technology Stack

- **Framework**: React 19 with TypeScript
- **State Management**: Redux Toolkit with RTK Query
- **Styling**: Tailwind CSS
- **Build Tool**: Vite
- **Testing**: Jest (unit), Playwright (E2E)
- **Code Quality**: ESLint, Prettier
- **UI Components**: Custom components with Radix UI

## Project Structure

```
ui/
├── src/
│   ├── app/              # Redux store and root configuration
│   │   ├── hooks.ts      # Typed Redux hooks
│   │   └── store.ts      # Store configuration
│   ├── components/       # Reusable React components
│   │   ├── common/       # Shared UI components
│   │   ├── flags/        # Flag-specific components
│   │   ├── segments/     # Segment-specific components
│   │   └── rules/        # Rule-specific components
│   ├── data/            # API integration and data layer
│   │   └── api/         # RTK Query API slices
│   ├── types/           # TypeScript type definitions
│   ├── utils/           # Utility functions and helpers
│   ├── App.tsx          # Root application component
│   └── main.tsx         # Application entry point
├── public/              # Static assets
├── tests/               # Test files
├── package.json         # Dependencies and scripts
├── tsconfig.json        # TypeScript configuration
├── vite.config.ts       # Vite build configuration
└── tailwind.config.js   # Tailwind CSS configuration
```

## Development Workflow

### Running the UI

```bash
# From repository root
mage ui:run

# Or from ui directory
npm run dev
```

The development server runs on port 5173 and proxies API requests to the backend on port 8080.

### Building for Production

```bash
# From repository root
mage ui:build

# Or from ui directory
npm run build
```

Production builds are embedded into the server binary during the main build process.

## Code Style Guidelines

### Component Structure

```tsx
// Use functional components with TypeScript
export default function FlagList({ namespace }: FlagListProps) {
  const { data, isLoading, error } = useListFlagsQuery({ namespace });

  if (isLoading) return <Loading />;
  if (error) return <ErrorMessage error={error} />;

  return (
    <div className="space-y-4">
      {data?.flags.map((flag) => (
        <FlagCard key={flag.key} flag={flag} />
      ))}
    </div>
  );
}
```

### State Management

Use Redux Toolkit with RTK Query for API calls:

```typescript
// data/api/flagsApi.ts
export const flagsApi = createApi({
  reducerPath: 'flagsApi',
  baseQuery: fetchBaseQuery({ baseUrl: '/api/v1' }),
  tagTypes: ['Flag'],
  endpoints: (builder) => ({
    listFlags: builder.query<FlagList, ListFlagsRequest>({
      query: ({ namespace }) => `namespaces/${namespace}/flags`,
      providesTags: ['Flag']
    })
  })
});
```

### TypeScript Best Practices

- Define interfaces for all props and API responses
- Use strict typing throughout the application
- Avoid `any` types unless absolutely necessary
- Export types from their domain modules

### Styling with Tailwind

```tsx
// Use Tailwind utility classes
<div className="flex items-center justify-between p-4 bg-white rounded-lg shadow">
  <h2 className="text-lg font-semibold text-gray-900">Flag Name</h2>
  <Switch enabled={enabled} onChange={handleToggle} />
</div>
```

### File Naming Conventions

- **Components**: PascalCase (`FlagList.tsx`, `SegmentEditor.tsx`)
- **Hooks**: camelCase with 'use' prefix (`useAuth.ts`, `useFlags.ts`)
- **Utils**: camelCase (`formatters.ts`, `validators.ts`)
- **API slices**: camelCase with 'Api' suffix (`flagsApi.ts`)
- **Types**: PascalCase (`Flag.ts`, `Segment.ts`)

## API Integration

### API Structure

The UI communicates with the backend through:

- **REST API**: Primary interface on `/api/v1` and `/api/v2`
- **WebSocket**: Real-time updates for flag changes (when enabled)

### Authentication

The UI supports multiple authentication methods:

- Basic authentication
- OIDC (OpenID Connect)
- Token-based authentication

Authentication state is managed in Redux and tokens are stored securely.

## Testing

### Unit Tests

```bash
# Run unit tests
npm run test

# Run with coverage
npm run test:coverage
```

### E2E Tests

```bash
# Run E2E tests (requires backend running)
npm run test:e2e
```

## Common Development Tasks

### Adding a New Feature

1. Create TypeScript types in `types/`
2. Add API endpoints in `data/api/`
3. Build React components in `components/`
4. Add routing if needed in `App.tsx`
5. Write tests for new functionality

### Modifying API Calls

1. Update the RTK Query slice in `data/api/`
2. Regenerate TypeScript types if needed
3. Update components using the API
4. Test with both mock and real backend

### Updating Styles

1. Use Tailwind classes for styling
2. Add custom styles sparingly in component files
3. Update `tailwind.config.js` for theme changes
4. Keep dark mode support in mind

## Build and Deployment

### Production Build Process

1. Vite bundles the application
2. Assets are optimized and minified
3. Build output goes to `dist/` directory
4. Server embeds the built assets

### Environment Variables

- `VITE_API_URL`: Backend API URL (defaults to proxy in dev)
- `VITE_ENABLE_TELEMETRY`: Enable usage telemetry
- `VITE_SENTRY_DSN`: Sentry error reporting DSN

## Performance Considerations

- Use React.memo for expensive components
- Implement virtualization for long lists
- Lazy load routes and heavy components
- Optimize bundle size with code splitting

## Debugging Tips

- Use React DevTools for component inspection
- Redux DevTools for state debugging
- Network tab for API call inspection
- Console for development logs
- Source maps are available in development

## Common Issues and Solutions

### CORS Issues

- Ensure proxy configuration in `vite.config.ts` is correct
- Check backend CORS settings if running separately

### State Management

- Clear Redux store on logout
- Handle optimistic updates carefully
- Invalidate cache tags after mutations

### Build Issues

- Clear `node_modules` and reinstall
- Check Node.js version (18+)
- Verify all dependencies are installed

## Important Notes

- The UI is embedded in the server binary for production
- Development runs as a separate process with proxy
- All API calls should handle loading and error states
- Maintain backward compatibility with v1 API
- Follow accessibility best practices (ARIA labels, keyboard navigation)
