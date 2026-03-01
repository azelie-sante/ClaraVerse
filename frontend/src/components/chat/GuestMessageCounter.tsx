interface GuestMessageCounterProps {
  used: number;
  max: number;
  onSignIn?: () => void;
}

/**
 * Small pill showing remaining free messages for guest users.
 * Displayed above the CommandCenter input area.
 * Shows a sign-in nudge once messages start running low.
 */
export const GuestMessageCounter: React.FC<GuestMessageCounterProps> = ({
  used,
  max,
  onSignIn,
}) => {
  const remaining = max - used;

  const getColor = () => {
    if (remaining >= 3) return '#10b981'; // green
    if (remaining === 2) return '#f59e0b'; // yellow
    return '#f97316'; // orange
  };

  return (
    <div
      style={{
        display: 'flex',
        flexDirection: 'column',
        alignItems: 'center',
        gap: '0.25rem',
        paddingBottom: '0.5rem',
      }}
    >
      <span
        style={{
          display: 'inline-flex',
          alignItems: 'center',
          gap: '0.375rem',
          padding: '0.25rem 0.75rem',
          borderRadius: '9999px',
          fontSize: '0.75rem',
          color: '#6b7280',
          background: 'rgba(0, 0, 0, 0.03)',
          border: '1px solid rgba(0, 0, 0, 0.06)',
        }}
      >
        <span
          style={{
            width: '6px',
            height: '6px',
            borderRadius: '50%',
            background: getColor(),
          }}
        />
        {remaining} of {max} free messages remaining
        {used > 0 && onSignIn && (
          <>
            {' · '}
            <button
              onClick={onSignIn}
              style={{
                background: 'none',
                border: 'none',
                padding: 0,
                color: '#6366f1',
                fontSize: '0.75rem',
                fontWeight: 500,
                cursor: 'pointer',
                textDecoration: 'underline',
                textUnderlineOffset: '2px',
              }}
            >
              Sign in to save chat
            </button>
          </>
        )}
      </span>
    </div>
  );
};
