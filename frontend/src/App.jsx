import { useState, useEffect, useRef } from 'react';
import PostFeed from './components/PostFeed';
import CreatePostModal from './components/CreatePostModal';

function App() {
  const [isModalOpen, setIsModalOpen] = useState(false);
  const [successMessage, setSuccessMessage] = useState('');
  const [newPost, setNewPost] = useState(null);
  const [searchQuery, setSearchQuery] = useState('');
  const [debouncedSearchQuery, setDebouncedSearchQuery] = useState('');
  const wsRef = useRef(null);

  // WebSocket connection for real-time updates
  useEffect(() => {
  let isMounted = true;
  let reconnectTimer = null;

  const connectWebSocket = () => {
    if (!isMounted) return;

    // Prevent duplicate connections (OPEN or CONNECTING)
    const existing = wsRef.current;
    if (
      existing &&
      (existing.readyState === WebSocket.OPEN ||
        existing.readyState === WebSocket.CONNECTING)
    ) {
      return;
    }
    // TODO. development mode only localhost.
    const ws = new WebSocket("ws://localhost:8080/api/v1/ws");
    wsRef.current = ws;

    ws.onopen = () => {
      console.log("WebSocket connected");
    };

    ws.onmessage = (event) => {
      try {
        const message = JSON.parse(event.data);
        if (message.type === "post_created" && message.data) {
          setNewPost(message.data);
        }
      } catch (err) {
        console.error("Failed to parse WebSocket msg:", err);
      }
    };

    ws.onerror = (err) => {
      console.error("WebSocket error:", err);
    };

    ws.onclose = () => {
      wsRef.current = null; 
      if (isMounted) {
        reconnectTimer = setTimeout(connectWebSocket, 3000);
      }
    };
  };

  connectWebSocket();

  return () => {
    isMounted = false;

    if (reconnectTimer) clearTimeout(reconnectTimer);

    const ws = wsRef.current;
    wsRef.current = null;

    if (
      ws &&
      (ws.readyState === WebSocket.OPEN ||
        ws.readyState === WebSocket.CONNECTING)
    ) {
      ws.close();
    }
  };
}, []);

  // Debounce search query by 300ms
  useEffect(() => {
    const timer = setTimeout(() => {
      setDebouncedSearchQuery(searchQuery);
    }, 300);

    return () => clearTimeout(timer);
  }, [searchQuery]);

  const handlePostCreated = (post) => {
    setNewPost(post);
    setSuccessMessage('Your post has been created successfully');
    setTimeout(() => setSuccessMessage(''), 3000);
  };

  return (
    <div className="min-h-screen bg-gray-100">
      {/* Success Toast */}
      {successMessage && (
        <div className="fixed top-4 left-1/2 -translate-x-1/2 z-50 bg-green-500 text-white px-6 py-3 rounded-lg shadow-lg whitespace-nowrap">
          {successMessage}
        </div>
      )}

      {/* Header */}
      <header className="sticky top-0 z-40 bg-white border-b shadow-sm">
        <div className="max-w-lg mx-auto px-4 py-3">
          <h1 className="text-xl font-bold bg-gradient-to-r from-purple-500 to-pink-500 bg-clip-text text-transparent">
            Instagram Lite
          </h1>
          <div className="mt-3 relative">
            <input
              type="text"
              value={searchQuery}
              onChange={(e) => setSearchQuery(e.target.value)}
              placeholder="Search by tags..."
              className="w-full pl-10 pr-4 py-2 border border-gray-300 rounded-lg focus:outline-none focus:ring-2 focus:ring-purple-500 focus:border-transparent"
            />
            <svg
              xmlns="http://www.w3.org/2000/svg"
              className="h-5 w-5 text-gray-400 absolute left-3 top-1/2 -translate-y-1/2"
              fill="none"
              viewBox="0 0 24 24"
              stroke="currentColor"
            >
              <path
                strokeLinecap="round"
                strokeLinejoin="round"
                strokeWidth={2}
                d="M21 21l-6-6m2-5a7 7 0 11-14 0 7 7 0 0114 0z"
              />
            </svg>
          </div>
        </div>
      </header>

      {/* Feed */}
      <main className="max-w-lg mx-auto px-4 py-6 pb-24">
        <PostFeed newPost={newPost} searchQuery={debouncedSearchQuery} />
      </main>

      {/* Floating Create Button */}
      <button
        onClick={() => setIsModalOpen(true)}
        className="fixed bottom-6 right-6 w-14 h-14 bg-gradient-to-r from-purple-500 to-pink-500 text-white rounded-full shadow-lg hover:shadow-xl hover:scale-105 transition-all duration-200 flex items-center justify-center"
      >
        <svg
          xmlns="http://www.w3.org/2000/svg"
          className="h-7 w-7"
          fill="none"
          viewBox="0 0 24 24"
          stroke="currentColor"
        >
          <path
            strokeLinecap="round"
            strokeLinejoin="round"
            strokeWidth={2}
            d="M12 4v16m8-8H4"
          />
        </svg>
      </button>

      {/* Create Post Modal */}
      <CreatePostModal
        isOpen={isModalOpen}
        onClose={() => setIsModalOpen(false)}
        onPostCreated={handlePostCreated}
      />
    </div>
  );
}

export default App;
