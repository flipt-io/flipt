/**
 * @jest-environment jsdom
 */
import { configureStore } from '@reduxjs/toolkit';
import { render, screen, fireEvent, act } from '@testing-library/react';
import { Provider } from 'react-redux';
import { MemoryRouter } from 'react-router';
import FlagTable from './FlagTable';

// ── Mock Radix Popover (portal-based, doesn't render in jsdom) ────────────
jest.mock('~/components/Popover', () => ({
  Popover: ({ children }: { children: React.ReactNode }) => <div>{children}</div>,
  PopoverTrigger: ({ children }: { children: React.ReactNode }) => <div>{children}</div>,
  PopoverContent: ({ children }: { children: React.ReactNode }) => <div>{children}</div>
}));

// ── Mock RTK Query hooks (actual import paths from FlagTable.tsx) ─────────
const mockFlags = [
  {
    key: 'alpha',
    name: 'Alpha',
    type: 'VARIANT_FLAG_TYPE',
    enabled: true,
    description: '',
    metadata: { team: 'backend', env: 'production' }
  },
  {
    key: 'beta',
    name: 'Beta',
    type: 'VARIANT_FLAG_TYPE',
    enabled: true,
    description: '',
    metadata: { team: 'frontend', env: 'production' }
  },
  {
    key: 'gamma',
    name: 'Gamma',
    type: 'VARIANT_FLAG_TYPE',
    enabled: true,
    description: '',
    metadata: { team: 'backend', env: 'staging' }
  }
];

jest.mock('~/app/flags/flagsApi', () => ({
  useListFlagsQuery: () => ({
    data: { flags: mockFlags },
    isLoading: false,
    error: null
  }),
  selectSorting: () => [],
  setSorting: (s: any) => ({ type: 'SET_SORTING', payload: s })
}));

jest.mock('~/app/flags/analyticsApi', () => ({
  useGetBatchFlagEvaluationCountQuery: () => ({ data: null })
}));

jest.mock('~/app/meta/metaSlice', () => ({
  selectInfo: () => ({ analytics: { enabled: false } })
}));

// ── Helpers ───────────────────────────────────────────────────────────────
const mockEnvironment = { key: 'default', name: 'default' };
const mockNamespace = { key: 'default', name: 'default', description: '' };

function renderTable() {
  const store = configureStore({
    reducer: {
      flags: (s = {}) => s,
      flagsTable: (s = { sorting: [] }) => s,
      meta: (s = { info: { analytics: { enabled: false } } }) => s
    }
  });
  return render(
    <MemoryRouter>
      <Provider store={store}>
        <FlagTable environment={mockEnvironment as any} namespace={mockNamespace as any} />
      </Provider>
    </MemoryRouter>
  );
}

// ── Fake timers for Searchbox debounce (defaults to 500ms) ────────────────
beforeEach(() => jest.useFakeTimers());
afterEach(() => jest.useRealTimers());

describe('FlagTable — metadata filter', () => {
  it('renders all flags with no filter active', () => {
    renderTable();
    expect(screen.getByText('Alpha')).toBeInTheDocument();
    expect(screen.getByText('Beta')).toBeInTheDocument();
    expect(screen.getByText('Gamma')).toBeInTheDocument();
  });

  it('renders Filter button in toolbar', () => {
    renderTable();
    expect(screen.getByRole('button', { name: /^filter$/i })).toBeInTheDocument();
  });

  it('adds a chip after applying a metadata filter', () => {
    renderTable();
    fireEvent.change(screen.getByPlaceholderText(/key/i), {
      target: { value: 'team' }
    });
    fireEvent.change(screen.getByPlaceholderText(/value/i), {
      target: { value: 'backend' }
    });
    fireEvent.click(screen.getByRole('button', { name: /add filter/i }));
    expect(screen.getByText(/team: backend/i)).toBeInTheDocument();
  });

  it('hides non-matching flags after filter is applied', () => {
    renderTable();
    fireEvent.change(screen.getByPlaceholderText(/key/i), {
      target: { value: 'team' }
    });
    fireEvent.change(screen.getByPlaceholderText(/value/i), {
      target: { value: 'backend' }
    });
    fireEvent.click(screen.getByRole('button', { name: /add filter/i }));
    expect(screen.queryByText('Beta')).not.toBeInTheDocument();
  });

  it('removes a chip when × is clicked and restores all flags', () => {
    renderTable();
    fireEvent.change(screen.getByPlaceholderText(/key/i), {
      target: { value: 'team' }
    });
    fireEvent.change(screen.getByPlaceholderText(/value/i), {
      target: { value: 'backend' }
    });
    fireEvent.click(screen.getByRole('button', { name: /add filter/i }));

    fireEvent.click(
      screen.getByRole('button', { name: /remove filter team:backend/i })
    );

    expect(screen.getByText('Beta')).toBeInTheDocument();
    expect(screen.queryByText(/team: backend/i)).not.toBeInTheDocument();
  });

  it('applies AND logic for two metadata filters', () => {
    renderTable();

    // Filter 1: team=backend
    fireEvent.change(screen.getByPlaceholderText(/key/i), {
      target: { value: 'team' }
    });
    fireEvent.change(screen.getByPlaceholderText(/value/i), {
      target: { value: 'backend' }
    });
    fireEvent.click(screen.getByRole('button', { name: /add filter/i }));

    // Filter 2: env=production
    fireEvent.change(screen.getByPlaceholderText(/key/i), {
      target: { value: 'env' }
    });
    fireEvent.change(screen.getByPlaceholderText(/value/i), {
      target: { value: 'production' }
    });
    fireEvent.click(screen.getByRole('button', { name: /add filter/i }));

    // Only alpha matches both filters
    expect(screen.getByText('Alpha')).toBeInTheDocument();
    expect(screen.queryByText('Beta')).not.toBeInTheDocument();
    expect(screen.queryByText('Gamma')).not.toBeInTheDocument();
  });

  it('shows "no flags matched" empty state when filter matches nothing', () => {
    renderTable();
    fireEvent.change(screen.getByPlaceholderText(/key/i), {
      target: { value: 'nonexistent' }
    });
    fireEvent.change(screen.getByPlaceholderText(/value/i), {
      target: { value: 'x' }
    });
    fireEvent.click(screen.getByRole('button', { name: /add filter/i }));

    expect(screen.getByText(/no flags matched/i)).toBeInTheDocument();
  });

  it('clears all filters when Clear all is clicked', () => {
    renderTable();
    fireEvent.change(screen.getByPlaceholderText(/key/i), {
      target: { value: 'team' }
    });
    fireEvent.change(screen.getByPlaceholderText(/value/i), {
      target: { value: 'backend' }
    });
    fireEvent.click(screen.getByRole('button', { name: /add filter/i }));

    fireEvent.click(screen.getByText('Clear all'));

    expect(screen.getByText('Beta')).toBeInTheDocument();
    expect(screen.queryByText(/team: backend/i)).not.toBeInTheDocument();
  });

  it('applies both text search and metadata filter simultaneously (AND)', () => {
    renderTable();

    // Text search: "alpha" — uniquely matches only the Alpha row
    fireEvent.change(screen.getByRole('searchbox'), {
      target: { value: 'alpha' }
    });
    act(() => {
      jest.advanceTimersByTime(600);
    }); // flush 500ms Searchbox debounce

    // Metadata filter: team=backend — alone would return [Alpha, Gamma]
    fireEvent.change(screen.getByPlaceholderText(/key/i), {
      target: { value: 'team' }
    });
    fireEvent.change(screen.getByPlaceholderText(/value/i), {
      target: { value: 'backend' }
    });
    fireEvent.click(screen.getByRole('button', { name: /add filter/i }));

    // Combined AND: text("alpha") ∩ metadata(team=backend) = [Alpha] only
    expect(screen.getByText('Alpha')).toBeInTheDocument();
    expect(screen.queryByText('Beta')).not.toBeInTheDocument();
    expect(screen.queryByText('Gamma')).not.toBeInTheDocument();
  });
});
