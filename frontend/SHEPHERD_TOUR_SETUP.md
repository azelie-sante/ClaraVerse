# Shepherd.js Tour Setup Guide

## ✅ Completed Setup

### 1. Installation
- ✅ Installed `shepherd.js` and `react-shepherd`
- ✅ Added custom CSS theme at `src/styles/shepherd-theme.css`
- ✅ Imported CSS in `src/App.tsx`

### 2. Infrastructure Created
- ✅ **TypeScript types**: `src/types/shepherd.d.ts`
- ✅ **Tour service**: `src/services/tourService.ts`
  - Manages tour completion state (localStorage)
  - Creates and caches tour instances
  - Provides helper methods (isTourCompleted, markTourCompleted, resetTours)

- ✅ **Custom hook**: `src/hooks/useShepherdTour.ts`
  - React hook for tour management
  - Supports auto-start and first-visit triggers

- ✅ **Tour button component**: `src/components/ui/TourButton.tsx`
  - Reusable button with icon/button variants
  - Shows tour status (completed/not completed)
  - Integrated with tour service

### 3. Tour Definitions
All tours created in `src/tours/` directory:

- ✅ **Dashboard Tour** (`dashboardTour.ts`)
  - Welcome message
  - Chat card highlight
  - Agents card highlight
  - Settings card highlight
  - User menu highlight

- ✅ **Chat Tour** (`chatTour.ts`)
  - Welcome to chat
  - Chat history sidebar
  - New chat button
  - Model selector
  - Message input
  - File upload
  - Chat options

- ✅ **Agents Tour** (`agentsTour.ts`)
  - Welcome to agents
  - Agents list
  - Create new agent button
  - Workflow canvas
  - Block palette
  - Block settings
  - Test agent button
  - Deploy agent button

### 4. Dashboard Integration
- ✅ Tour button added to top-right header
- ✅ Data attributes added to app cards:
  - `data-tour="chat-card"`
  - `data-tour="agents-card"`
  - `data-tour="settings-card"`
  - `data-tour="user-menu"`

---

## 🔧 Next Steps: Complete Chat & Agents Integration

### Chat Page Integration

You need to add `data-tour` attributes to these elements in `/src/pages/Chat.tsx`:

```tsx
// 1. Chat sidebar (likely in the Sidebar component props)
<Sidebar data-tour="chat-sidebar" {...otherProps} />

// 2. New chat button (search for Plus icon or "New Chat")
<button data-tour="new-chat">...</button>

// 3. Model selector (search for model dropdown/selector)
<div data-tour="model-selector">...</div>

// 4. Message input textarea
<textarea data-tour="message-input">...</textarea>

// 5. File upload button (search for attach/upload icon)
<button data-tour="file-upload">...</button>

// 6. Chat options menu (likely settings/options in header)
<div data-tour="chat-options">...</div>
```

Then add the tour button in the Chat page header:

```tsx
import { TourButton } from '@/components/ui';
import { chatTourSteps } from '@/tours';

// In the header/toolbar area:
<TourButton tourName="chat" steps={chatTourSteps} variant="icon" />
```

### Agents Page Integration

You need to add `data-tour` attributes to these elements in `/src/pages/Agents.tsx`:

```tsx
// 1. Agents list sidebar or panel
<div data-tour="agents-list">...</div>

// 2. Create agent button
<button data-tour="create-agent">...</button>

// 3. Workflow canvas (main working area)
<div data-tour="workflow-canvas">...</div>

// 4. Block palette (drag-and-drop blocks)
<div data-tour="block-palette">...</div>

// 5. Block settings panel
<div data-tour="block-settings">...</div>

// 6. Test button
<button data-tour="test-agent">...</button>

// 7. Deploy button
<button data-tour="deploy-agent">...</button>
```

Then add the tour button in the Agents page header:

```tsx
import { TourButton } from '@/components/ui';
import { agentsTourSteps } from '@/tours';

// In the header/toolbar area:
<TourButton tourName="agents" steps={agentsTourSteps} variant="icon" />
```

---

## 📝 Usage Examples

### Trigger a Tour Programmatically

```tsx
import { tourService } from '@/services/tourService';
import { chatTourSteps } from '@/tours';

// Start a tour
tourService.startTour('chat', chatTourSteps);

// Check if completed
const isCompleted = tourService.isTourCompleted('chat');

// Reset tour
tourService.resetTours('chat');
```

### Use the Hook

```tsx
import { useShepherdTour } from '@/hooks/useShepherdTour';
import { dashboardTourSteps } from '@/tours';

function MyComponent() {
  const { startTour, isCompleted, isActive } = useShepherdTour({
    tourName: 'dashboard',
    steps: dashboardTourSteps,
    startOnFirstVisit: true, // Auto-start on first visit
  });

  return (
    <button onClick={startTour}>
      {isCompleted ? 'Restart Tour' : 'Take Tour'}
    </button>
  );
}
```

### Customize Tour Steps

Edit tour files in `src/tours/` to:
- Change text and titles
- Adjust element positions (`on: 'top' | 'bottom' | 'left' | 'right'`)
- Add/remove steps
- Customize buttons

---

## 🎨 Styling

The custom theme is in `src/styles/shepherd-theme.css` with:
- Modern, clean design
- Dark mode support
- Smooth animations
- Mobile responsive
- Matches your app's design system

To customize colors/spacing, edit the CSS file.

---

## 🔍 Finding Elements

To find where to add `data-tour` attributes:

1. **Search for component names**: "ModelSelector", "FileUpload", etc.
2. **Search for icon imports**: `Plus`, `Upload`, `Settings`, etc.
3. **Use browser DevTools**: Inspect elements while the page is running
4. **Check component props**: Many components may already accept className or data-* props

---

## ✨ Features

- ✅ Persistent tour completion state (localStorage)
- ✅ Auto-start tours on first visit
- ✅ Skip/cancel tour anytime
- ✅ Back/Next navigation
- ✅ Keyboard navigation (arrow keys, ESC to exit)
- ✅ Modal overlay with highlighted elements
- ✅ Mobile responsive
- ✅ Dark mode support
- ✅ Tour reset functionality

---

## 🐛 Troubleshooting

**Tour not showing?**
- Ensure the target element exists in DOM
- Check `data-tour` attribute is correctly applied
- Verify element is visible (not hidden or in a collapsed panel)
- Use `display: none` for elements that should be skipped

**Element not highlighting correctly?**
- Adjust the `on` position: try 'auto' for automatic placement
- Increase `modalOverlayOpeningPadding` for more space
- Ensure element has proper dimensions

**Tour showing at wrong time?**
- Check `startOnFirstVisit` option
- Verify `tourService.isTourCompleted()` logic
- Clear localStorage to test first-visit behavior

---

## 📦 Project Structure

```
frontend/src/
├── tours/
│   ├── index.ts              # Export all tours
│   ├── dashboardTour.ts      # Dashboard tour steps
│   ├── chatTour.ts           # Chat tour steps
│   └── agentsTour.ts         # Agents tour steps
├── services/
│   └── tourService.ts        # Tour management service
├── hooks/
│   └── useShepherdTour.ts    # React hook for tours
├── components/ui/
│   └── TourButton.tsx        # Reusable tour button
├── types/
│   └── shepherd.d.ts         # TypeScript definitions
└── styles/
    └── shepherd-theme.css    # Custom tour styling
```

---

## 🚀 Testing

1. Clear localStorage: `localStorage.clear()`
2. Refresh the page
3. Tours should auto-start on first visit (if configured)
4. Click tour buttons to manually start tours
5. Test all navigation (Next, Back, Skip)
6. Verify completion state persists across page reloads

---

## 📊 Analytics & Tracking

### Features

The tour system now includes comprehensive analytics tracking:

- **Auto-start for first-time desktop users**: Tours automatically start for new users on desktop devices
- **Event tracking**: Tracks tour starts, completions, skips, and step views
- **Persistent storage**: All analytics data stored in localStorage
- **Statistics**: Calculate completion rates and user engagement metrics

### How It Works

1. **First-Time Detection**:
   - Checks if user is on desktop (≥1024px width)
   - Checks if user has visited before
   - Auto-starts tour if both conditions are met

2. **Analytics Events**:
   - `started`: When a tour begins
   - `completed`: When user finishes the entire tour
   - `skipped`: When user cancels/closes tour early
   - `step_viewed`: Each time a step is shown

### Accessing Analytics Data

#### Console Methods

Open browser console and use these commands:

```javascript
// Get all analytics events
tourService.getTourAnalytics();

// Get analytics for specific tour
tourService.getTourAnalytics('dashboard');

// Get statistics
tourService.getTourStats('dashboard');
// Returns: { started: 10, completed: 8, skipped: 2, completionRate: 80 }

// Check if first-time user
tourService.isFirstTimeUser();

// Reset first-time user flag (for testing)
tourService.resetFirstTimeUser();

// Clear all analytics
tourService.clearAnalytics();
```

#### Programmatic Access

```typescript
import { tourService } from '@/services/tourService';

// Get dashboard tour statistics
const stats = tourService.getTourStats('dashboard');
console.log(`Completion rate: ${stats.completionRate}%`);
console.log(`Skipped: ${stats.skipped} times`);

// Get all events
const events = tourService.getTourAnalytics();
events.forEach(event => {
  console.log(`${event.tourName}: ${event.action} at ${event.timestamp}`);
});
```

### LocalStorage Keys

Analytics data is stored in these localStorage keys:

- `shepherd-tour-completed`: Array of completed tour names
- `shepherd-tour-analytics`: Array of all tour events
- `clara-first-time-user`: Flag indicating if user has visited before

### Testing First-Time User Experience

To test the auto-start behavior:

```javascript
// 1. Clear first-time user flag
tourService.resetFirstTimeUser();

// 2. Clear tour completion
tourService.resetTours('dashboard');

// 3. Reload the page
location.reload();
```

Or clear all tour data at once:

```javascript
// Clear everything
localStorage.removeItem('shepherd-tour-completed');
localStorage.removeItem('shepherd-tour-analytics');
localStorage.removeItem('clara-first-time-user');
location.reload();
```

### Viewing Analytics Dashboard

To create a simple analytics view, add this to an admin panel:

```tsx
import { tourService, type TourStats } from '@/services/tourService';
import { useEffect, useState } from 'react';

function TourAnalyticsDashboard() {
  const [stats, setStats] = useState<TourStats>({
    started: 0,
    completed: 0,
    skipped: 0,
    completionRate: 0,
  });

  useEffect(() => {
    setStats(tourService.getTourStats('dashboard'));
  }, []);

  return (
    <div>
      <h2>Dashboard Tour Analytics</h2>
      <p>Started: {stats.started}</p>
      <p>Completed: {stats.completed}</p>
      <p>Skipped: {stats.skipped}</p>
      <p>Completion Rate: {stats.completionRate.toFixed(1)}%</p>
    </div>
  );
}
```

### Integrating with Analytics Services

To send events to Google Analytics, Mixpanel, or other services, modify the `trackAnalytics` method in `tourService.ts`:

```typescript
private trackAnalytics(event: TourAnalytics): void {
  // Store locally
  const analytics = localStorage.getItem(this.TOUR_ANALYTICS_KEY);
  const events: TourAnalytics[] = analytics ? JSON.parse(analytics) : [];
  events.push(event);
  localStorage.setItem(this.TOUR_ANALYTICS_KEY, JSON.stringify(events));

  // Send to external analytics service
  if (typeof gtag !== 'undefined') {
    gtag('event', event.action, {
      event_category: 'tour',
      event_label: event.tourName,
      value: event.stepId,
    });
  }
}
```

---

Happy touring! 🎉
