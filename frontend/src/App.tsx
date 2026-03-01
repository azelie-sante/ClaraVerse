import { useEffect } from 'react';
import { AppRouter } from '@/routes';
import { ToastContainer } from '@/components/design-system';
import { UpgradePromptModal, WelcomePopupModal } from '@/components/subscription';

function App() {
  // Apply Christmas theme during Dec 15 - Jan 1
  useEffect(() => {
    const now = new Date();
    const month = now.getMonth(); // 0-11
    const day = now.getDate();

    // Dec 15 - Jan 1 (month 11 = Dec, month 0 = Jan)
    const isChristmasSeason = (month === 11 && day >= 15) || (month === 0 && day <= 1);

    if (isChristmasSeason) {
      document.documentElement.setAttribute('data-theme', 'christmas');
    }

    return () => {
      document.documentElement.removeAttribute('data-theme');
    };
  }, []);

  return (
    <>
      <AppRouter />
      <ToastContainer />
      <UpgradePromptModal />
      <WelcomePopupModal />
    </>
  );
}

export default App;
