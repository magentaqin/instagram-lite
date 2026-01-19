import { useState, useEffect, useRef, useCallback } from 'react';

function PostFeed({ newPost, searchQuery }) {
  const [posts, setPosts] = useState([]);
  const [cursor, setCursor] = useState(null);
  const [loading, setLoading] = useState(false);
  const [hasMore, setHasMore] = useState(true);
  const [error, setError] = useState('');
  // To get the sentinel DOM node
  const observerRef = useRef();

  // Fetch the post list
  const fetchPosts = useCallback(async (isNewSearch = false) => {
    if (loading || (!hasMore && !isNewSearch)) return;

    setLoading(true);

    try {
      const params = new URLSearchParams();
      params.append('limit', '10');

      if (searchQuery?.trim()) {
        params.append('tag', searchQuery.trim());
      }

      // Use cursor for pagination, but not for new searches
      if (!isNewSearch && cursor) {
        params.append('cursor', cursor);
      }

      const res = await fetch(`/api/v1/posts?${params.toString()}`);
      const data = await res.json();

      if (!res.ok) {
        throw new Error(data.error || 'Failed to fetch posts');
      }

      // if search query changes, reset it. or, we just append them.
      if (isNewSearch) {
        setPosts(data.items || []);
      } else {
        setPosts((prev) => [...prev, ...(data.items || [])]);
      }

      setCursor(data.next_cursor || null);
      setHasMore(data.has_more || false);
    } catch (err) {
      console.error('Failed to fetch posts:', err);
      setError('Failed to load posts');
      setTimeout(() => setError(''), 3000);
    } finally {
      setLoading(false);
    }
  }, [cursor, loading, hasMore, searchQuery]);

  const lastPostRef = useCallback(
    (node) => {
      if (loading) return;
      // disconnect previous observer to avoid duplicated trigger
      if (observerRef.current) observerRef.current.disconnect();
      // create observerRef
      observerRef.current = new IntersectionObserver((entries) => {
        // if it's in our viewport, and we have more posts, fetchPosts
        if (entries[0].isIntersecting && hasMore) {
          fetchPosts();
        }
      });

      if (node) observerRef.current.observe(node);
    },
    [loading, hasMore, fetchPosts]
  );

  // Listen for new post event
  useEffect(() => {
    if (newPost) {
      setPosts((prev) => [newPost, ...prev]);
    }
  }, [newPost]);

  // Fetch posts on mount and when search query changes
  useEffect(() => {
    // Reset and fetch with new search
    setCursor(null);
    setHasMore(true);
    fetchPosts(true);
  }, [searchQuery]);

  // Empty state: no posts yet, or no matches for the current tag search.
  if (posts.length === 0 && !loading) {
    return (
      <div className="flex flex-col items-center justify-center py-20 text-gray-500">
        <svg
          xmlns="http://www.w3.org/2000/svg"
          className="h-16 w-16 mb-4"
          fill="none"
          viewBox="0 0 24 24"
          stroke="currentColor"
        >
          <path
            strokeLinecap="round"
            strokeLinejoin="round"
            strokeWidth={1.5}
            d="M4 16l4.586-4.586a2 2 0 012.828 0L16 16m-2-2l1.586-1.586a2 2 0 012.828 0L20 14m-6-6h.01M6 20h12a2 2 0 002-2V6a2 2 0 00-2-2H6a2 2 0 00-2 2v12a2 2 0 002 2z"
          />
        </svg>
        <p className="text-lg">{searchQuery?.trim() ? 'No matching posts' : 'No posts yet'}</p>
        <p className="text-sm">{searchQuery?.trim() ? 'Try a different tag search' : 'Be the first to share a moment!'}</p>
      </div>
    );
  }

  return (
    <div className="space-y-6">
      {posts.map((post, index) => {
        const isLast = index === posts.length - 1;
        return (
          <div
            key={post.id}
            ref={isLast ? lastPostRef : null}
            className="bg-white rounded-lg shadow overflow-hidden"
          >
            <div className="flex items-center px-4 py-3">
              <div className="w-8 h-8 rounded-full bg-gradient-to-br from-purple-500 to-pink-500 flex items-center justify-center text-white text-sm font-semibold">
                U
              </div>
              <span className="ml-3 font-semibold text-sm">user</span>
            </div>
            <img
              src={post.image_url}
              alt=""
              className="w-full aspect-square object-cover"
            />
            <div className="p-4">
              <div className="flex items-center gap-4 mb-2">
                <button className="hover:text-red-500 transition-colors">
                  <svg xmlns="http://www.w3.org/2000/svg" className="h-6 w-6" fill="none" viewBox="0 0 24 24" stroke="currentColor">
                    <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M4.318 6.318a4.5 4.5 0 000 6.364L12 20.364l7.682-7.682a4.5 4.5 0 00-6.364-6.364L12 7.636l-1.318-1.318a4.5 4.5 0 00-6.364 0z" />
                  </svg>
                </button>
                <button className="hover:text-gray-600 transition-colors">
                  <svg xmlns="http://www.w3.org/2000/svg" className="h-6 w-6" fill="none" viewBox="0 0 24 24" stroke="currentColor">
                    <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M8 12h.01M12 12h.01M16 12h.01M21 12c0 4.418-4.03 8-9 8a9.863 9.863 0 01-4.255-.949L3 20l1.395-3.72C3.512 15.042 3 13.574 3 12c0-4.418 4.03-8 9-8s9 3.582 9 8z" />
                  </svg>
                </button>
              </div>
              {post.title && (
                <p className="font-semibold text-sm mb-1">{post.title}</p>
              )}
              {post.tags && post.tags.length > 0 && (
                <div className="flex flex-wrap gap-1 mt-2">
                  {post.tags.map((tag) => (
                    <span
                      key={tag}
                      className="text-xs text-purple-600 hover:text-purple-800 cursor-pointer"
                    >
                      #{tag}
                    </span>
                  ))}
                </div>
              )}
              <p className="text-xs text-gray-500 mt-2">
                {new Date(post.created_at).toLocaleDateString()}
              </p>
            </div>
          </div>
        );
      })}

      {loading && (
        <div className="flex justify-center py-4">
          <div className="w-8 h-8 border-4 border-purple-500 border-t-transparent rounded-full animate-spin"></div>
        </div>
      )}

      {error && (
        <div className="fixed bottom-20 left-1/2 -translate-x-1/2 bg-red-100 text-red-700 px-4 py-2 rounded-lg shadow-md text-sm whitespace-nowrap">
          {error}
        </div>
      )}
    </div>
  );
}

export default PostFeed;
