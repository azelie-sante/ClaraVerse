export const About = () => {
  return (
    <div className="space-y-6">
      <h1 className="text-4xl font-bold text-gray-900">About</h1>
      <div className="prose prose-lg">
        <p className="text-gray-600">This is a production-ready React application featuring:</p>
        <ul className="list-disc list-inside space-y-2 text-gray-600">
          <li>React 18 (LTS) with TypeScript</li>
          <li>Vite for fast development and optimized builds</li>
          <li>Tailwind CSS for styling</li>
          <li>React Router v6 for navigation</li>
          <li>Zustand for state management</li>
          <li>ESLint and Prettier for code quality</li>
        </ul>
      </div>
    </div>
  );
};
