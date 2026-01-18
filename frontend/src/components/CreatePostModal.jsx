import { useState } from 'react';
import ImageUpload from './ImageUpload';

function CreatePostModal({ isOpen, onClose, onPostCreated }) {
  const [imageUrl, setImageUrl] = useState('');
  const [uploading, setUploading] = useState(false);
  const [posting, setPosting] = useState(false);
  const [error, setError] = useState('');
  const [title, setTitle] = useState('');
  const [tagInput, setTagInput] = useState('');
  const [tags, setTags] = useState([]);

  // Reset State
  const resetState = () => {
    setImageUrl('');
    setUploading(false);
    setPosting(false);
    setError('');
    setTitle('');
    setTagInput('');
    setTags([]);
  };

  // Close modal
  const handleClose = () => {
    resetState();
    onClose();
  };

  // Upload image to backend and image url to render the preview area
  const handleFileSelect = async (file) => {
    setUploading(true);
    setError('');

    const formData = new FormData();
    formData.append('file', file);

    try {
      const res = await fetch('/api/v1/upload', {
        method: 'POST',
        body: formData,
      });

      const data = await res.json();

      if (!res.ok) {
        throw new Error(data.error || 'Upload failed');
      }
      // Use the response.data.image_url as the image preview src
      setImageUrl(data.image_url);
    } catch (err) {
      setError(err.message);
    } finally {
      setUploading(false);
    }
  };

  // Remove image handler
  const handleRemove = () => {
    setImageUrl('');
    setError('');
    setTitle('');
    setTagInput('');
    setTags([]);
  };

  // Add tag handler
  const handleAddTag = () => {
    const trimmedTag = tagInput.trim();
    if (trimmedTag.length > 16) {
      setError('Tag must be less than 16 characters');
      return;
    }
    if (trimmedTag && tags.length < 10 && !tags.includes(trimmedTag)) {
      setTags([...tags, trimmedTag]);
      setTagInput('');
      setError('');
    }
  };

  // Listener for keydown
  const handleTagKeyDown = (e) => {
    if (e.key === 'Enter') {
      // Prevent Enter from submitting the form; treat it only as "add tag".
      e.preventDefault();
      handleAddTag();
    }
  };

  // Remove tag
  const handleRemoveTag = (tagToRemove) => {
    setTags(tags.filter((tag) => tag !== tagToRemove));
  };

  // Validate form and create post
  const handleSubmit = async () => {
    if (!imageUrl) return;

    if (title.trim().length > 120) {
      setError('Title must be 120 characters or less');
      return;
    }

    // Set Loading state
    setPosting(true);
    // Clear error in previous validation
    setError('');

    try {
      const res = await fetch('/api/v1/posts', {
        method: 'POST',
        headers: {
          'Content-Type': 'application/json',
        },
        body: JSON.stringify({ image_url: imageUrl, title, tags }),
      });

      const data = await res.json();

      if (!res.ok) {
        throw new Error(data.error || 'Failed to create post');
      }

      resetState();
      onPostCreated(data);
      onClose();
    } catch (err) {
      setError(err.message);
    } finally {
      setPosting(false);
    }
  };

  // Here, we choose the easy way mount/unmount component instead of reaching out to the CSS way, because:
  // 1.it doesn't have animation 2.Css way has issues like scroll-through...
  if (!isOpen) return null;

  return (
    <div className="fixed inset-0 z-50 flex items-center justify-center">
      <div
        className="absolute inset-0 bg-black/50"
        onClick={handleClose}
      ></div>

      <div className="relative bg-white rounded-xl shadow-xl w-full max-w-md mx-4 overflow-hidden">
        <div className="flex items-center justify-between p-4 border-b">
          <h2 className="text-lg font-semibold">Create Post</h2>
          <button
            onClick={handleClose}
            className="p-1 hover:bg-gray-100 rounded-full transition-colors"
          >
            <svg
              xmlns="http://www.w3.org/2000/svg"
              className="h-6 w-6 text-gray-500"
              fill="none"
              viewBox="0 0 24 24"
              stroke="currentColor"
            >
              <path
                strokeLinecap="round"
                strokeLinejoin="round"
                strokeWidth={2}
                d="M6 18L18 6M6 6l12 12"
              />
            </svg>
          </button>
        </div>

        <div className="p-4">
          <ImageUpload
            imageUrl={imageUrl}
            uploading={uploading}
            onFileSelect={handleFileSelect}
            onRemove={handleRemove}
          />

          {imageUrl && (
            <div className="mt-4 space-y-4">
              <div>
                <label className="block text-sm font-medium text-gray-700 mb-1">
                  Title
                </label>
                <input
                  type="text"
                  value={title}
                  onChange={(e) => setTitle(e.target.value)}
                  placeholder="Enter post title"
                  className="w-full px-3 py-2 border border-gray-300 rounded-lg focus:outline-none focus:ring-2 focus:ring-purple-500 focus:border-transparent"
                />
              </div>

              <div>
                <label className="block text-sm font-medium text-gray-700 mb-1">
                  Tags ({tags.length}/10)
                </label>
                <div className="flex gap-2">
                  <input
                    type="text"
                    value={tagInput}
                    onChange={(e) => setTagInput(e.target.value)}
                    onKeyDown={handleTagKeyDown}
                    placeholder="Add a tag"
                    disabled={tags.length >= 10}
                    className="flex-1 px-3 py-2 border border-gray-300 rounded-lg focus:outline-none focus:ring-2 focus:ring-purple-500 focus:border-transparent disabled:bg-gray-100"
                  />
                  <button
                    type="button"
                    onClick={handleAddTag}
                    disabled={tags.length >= 10 || !tagInput.trim()}
                    className="px-4 py-2 bg-gray-100 text-gray-700 rounded-lg hover:bg-gray-200 transition-colors disabled:opacity-50 disabled:cursor-not-allowed"
                  >
                    Add
                  </button>
                </div>
                {tags.length > 0 && (
                  <div className="flex flex-wrap gap-2 mt-2">
                    {tags.map((tag) => (
                      <span
                        key={tag}
                        className="inline-flex items-center gap-1 px-2 py-1 bg-purple-100 text-purple-700 text-sm rounded-full"
                      >
                        #{tag}
                        <button
                          type="button"
                          onClick={() => handleRemoveTag(tag)}
                          className="hover:text-purple-900"
                        >
                          <svg
                            xmlns="http://www.w3.org/2000/svg"
                            className="h-3 w-3"
                            fill="none"
                            viewBox="0 0 24 24"
                            stroke="currentColor"
                          >
                            <path
                              strokeLinecap="round"
                              strokeLinejoin="round"
                              strokeWidth={2}
                              d="M6 18L18 6M6 6l12 12"
                            />
                          </svg>
                        </button>
                      </span>
                    ))}
                  </div>
                )}
              </div>

              <button
                onClick={handleSubmit}
                disabled={posting || !title.trim()}
                className="w-full py-3 px-6 bg-gradient-to-r from-purple-500 to-pink-500 text-white font-semibold rounded-lg hover:from-purple-600 hover:to-pink-600 transition-all duration-200 disabled:opacity-50 disabled:cursor-not-allowed"
              >
                {posting ? 'Posting...' : 'Create'}
              </button>
            </div>
          )}

          {error && <p className="mt-3 text-sm text-red-500">{error}</p>}
        </div>
      </div>
    </div>
  );
}

export default CreatePostModal;
