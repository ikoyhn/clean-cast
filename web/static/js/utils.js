// Clean Cast Admin Panel - Utility Functions

/**
 * Format duration from seconds to human-readable format
 */
function formatDuration(seconds) {
  if (!seconds || seconds === 0) return '0:00';

  const hours = Math.floor(seconds / 3600);
  const minutes = Math.floor((seconds % 3600) / 60);
  const secs = Math.floor(seconds % 60);

  if (hours > 0) {
    return `${hours}:${String(minutes).padStart(2, '0')}:${String(secs).padStart(2, '0')}`;
  }
  return `${minutes}:${String(secs).padStart(2, '0')}`;
}

/**
 * Format bytes to human-readable size
 */
function formatBytes(bytes, decimals = 2) {
  if (bytes === 0) return '0 Bytes';

  const k = 1024;
  const dm = decimals < 0 ? 0 : decimals;
  const sizes = ['Bytes', 'KB', 'MB', 'GB', 'TB'];
  const i = Math.floor(Math.log(bytes) / Math.log(k));

  return parseFloat((bytes / Math.pow(k, i)).toFixed(dm)) + ' ' + sizes[i];
}

/**
 * Format date to human-readable format
 */
function formatDate(dateString) {
  if (!dateString) return 'N/A';

  const date = new Date(dateString);
  const now = new Date();
  const diffMs = now - date;
  const diffDays = Math.floor(diffMs / (1000 * 60 * 60 * 24));

  if (diffDays === 0) {
    const diffHours = Math.floor(diffMs / (1000 * 60 * 60));
    if (diffHours === 0) {
      const diffMinutes = Math.floor(diffMs / (1000 * 60));
      if (diffMinutes === 0) return 'Just now';
      return `${diffMinutes} minute${diffMinutes > 1 ? 's' : ''} ago`;
    }
    return `${diffHours} hour${diffHours > 1 ? 's' : ''} ago`;
  }

  if (diffDays === 1) return 'Yesterday';
  if (diffDays < 7) return `${diffDays} days ago`;
  if (diffDays < 30) return `${Math.floor(diffDays / 7)} week${Math.floor(diffDays / 7) > 1 ? 's' : ''} ago`;

  return date.toLocaleDateString('en-US', {
    year: 'numeric',
    month: 'short',
    day: 'numeric'
  });
}

/**
 * Format number with commas
 */
function formatNumber(num) {
  if (num === null || num === undefined) return '0';
  return num.toString().replace(/\B(?=(\d{3})+(?!\d))/g, ',');
}

/**
 * Debounce function to limit rate of function calls
 */
function debounce(func, wait) {
  let timeout;
  return function executedFunction(...args) {
    const later = () => {
      clearTimeout(timeout);
      func(...args);
    };
    clearTimeout(timeout);
    timeout = setTimeout(later, wait);
  };
}

/**
 * Show toast notification
 */
function showToast(message, type = 'info', duration = 3000) {
  const toast = document.createElement('div');
  toast.className = `alert alert-${type} toast`;
  toast.style.cssText = `
    position: fixed;
    top: 20px;
    right: 20px;
    z-index: 9999;
    min-width: 300px;
    max-width: 500px;
    animation: slideInRight 0.3s ease;
  `;
  toast.textContent = message;

  document.body.appendChild(toast);

  setTimeout(() => {
    toast.style.animation = 'slideOutRight 0.3s ease';
    setTimeout(() => {
      document.body.removeChild(toast);
    }, 300);
  }, duration);
}

/**
 * Show loading indicator
 */
function showLoading(containerId) {
  const container = document.getElementById(containerId);
  if (!container) return;

  container.innerHTML = `
    <div class="flex justify-center items-center py-12">
      <div class="spinner"></div>
    </div>
  `;
}

/**
 * Hide loading and show content
 */
function hideLoading(containerId, content) {
  const container = document.getElementById(containerId);
  if (!container) return;

  container.innerHTML = content;
}

/**
 * Confirm dialog
 */
function confirmDialog(message, onConfirm, onCancel) {
  const overlay = document.createElement('div');
  overlay.className = 'modal-overlay';
  overlay.innerHTML = `
    <div class="modal">
      <div class="modal-header">
        <h3 class="text-lg font-semibold">Confirm Action</h3>
      </div>
      <div class="modal-body">
        <p>${message}</p>
      </div>
      <div class="modal-footer">
        <button class="btn btn-secondary" id="cancelBtn">Cancel</button>
        <button class="btn btn-danger" id="confirmBtn">Confirm</button>
      </div>
    </div>
  `;

  document.body.appendChild(overlay);

  document.getElementById('confirmBtn').addEventListener('click', () => {
    document.body.removeChild(overlay);
    if (onConfirm) onConfirm();
  });

  document.getElementById('cancelBtn').addEventListener('click', () => {
    document.body.removeChild(overlay);
    if (onCancel) onCancel();
  });

  overlay.addEventListener('click', (e) => {
    if (e.target === overlay) {
      document.body.removeChild(overlay);
      if (onCancel) onCancel();
    }
  });
}

/**
 * Show modal dialog
 */
function showModal(title, content, buttons = []) {
  const overlay = document.createElement('div');
  overlay.className = 'modal-overlay';
  overlay.id = 'customModal';

  let buttonsHtml = '';
  buttons.forEach(btn => {
    buttonsHtml += `<button class="btn btn-${btn.type || 'primary'}" id="${btn.id}">${btn.text}</button>`;
  });

  overlay.innerHTML = `
    <div class="modal">
      <div class="modal-header">
        <h3 class="text-lg font-semibold">${title}</h3>
        <button class="text-gray-400 hover:text-white" id="closeModal">
          <i class="fas fa-times"></i>
        </button>
      </div>
      <div class="modal-body">
        ${content}
      </div>
      <div class="modal-footer">
        ${buttonsHtml}
      </div>
    </div>
  `;

  document.body.appendChild(overlay);

  document.getElementById('closeModal').addEventListener('click', () => {
    document.body.removeChild(overlay);
  });

  overlay.addEventListener('click', (e) => {
    if (e.target === overlay) {
      document.body.removeChild(overlay);
    }
  });

  return overlay;
}

/**
 * Close modal
 */
function closeModal() {
  const modal = document.getElementById('customModal');
  if (modal) {
    document.body.removeChild(modal);
  }
}

/**
 * Copy to clipboard
 */
async function copyToClipboard(text) {
  try {
    await navigator.clipboard.writeText(text);
    showToast('Copied to clipboard!', 'success', 2000);
    return true;
  } catch (err) {
    console.error('Failed to copy:', err);
    showToast('Failed to copy to clipboard', 'error', 2000);
    return false;
  }
}

/**
 * Truncate text
 */
function truncateText(text, maxLength) {
  if (!text || text.length <= maxLength) return text;
  return text.substring(0, maxLength) + '...';
}

/**
 * Get query parameter from URL
 */
function getQueryParam(param) {
  const urlParams = new URLSearchParams(window.location.search);
  return urlParams.get(param);
}

/**
 * Update query parameter in URL
 */
function updateQueryParam(param, value) {
  const url = new URL(window.location);
  url.searchParams.set(param, value);
  window.history.pushState({}, '', url);
}

/**
 * Parse error response
 */
function parseError(error) {
  if (error.response && error.response.data) {
    if (error.response.data.error) return error.response.data.error;
    if (error.response.data.message) return error.response.data.message;
  }
  if (error.message) return error.message;
  return 'An unknown error occurred';
}

/**
 * Validate YouTube URL/ID
 */
function validateYoutubeId(input) {
  // Check if it's already a valid ID (11 characters)
  if (/^[a-zA-Z0-9_-]{11}$/.test(input)) {
    return input;
  }

  // Try to extract ID from URL
  const patterns = [
    /(?:youtube\.com\/watch\?v=|youtu\.be\/|youtube\.com\/embed\/)([a-zA-Z0-9_-]{11})/,
    /youtube\.com\/playlist\?list=([a-zA-Z0-9_-]+)/,
    /youtube\.com\/channel\/([a-zA-Z0-9_-]+)/,
    /youtube\.com\/@([a-zA-Z0-9_-]+)/
  ];

  for (const pattern of patterns) {
    const match = input.match(pattern);
    if (match) return match[1];
  }

  return null;
}

/**
 * Format podcast type badge
 */
function getPodcastTypeBadge(type) {
  const badges = {
    'PLAYLIST': '<span class="badge badge-primary">Playlist</span>',
    'CHANNEL': '<span class="badge badge-info">Channel</span>'
  };
  return badges[type] || '<span class="badge">Unknown</span>';
}

/**
 * Format status badge
 */
function getStatusBadge(status) {
  const badges = {
    'completed': '<span class="badge badge-success">Completed</span>',
    'pending': '<span class="badge badge-warning">Pending</span>',
    'running': '<span class="badge badge-primary">Running</span>',
    'failed': '<span class="badge badge-danger">Failed</span>'
  };
  return badges[status] || '<span class="badge">Unknown</span>';
}

/**
 * Sanitize HTML to prevent XSS
 */
function sanitizeHTML(html) {
  const temp = document.createElement('div');
  temp.textContent = html;
  return temp.innerHTML;
}

/**
 * Create pagination HTML
 */
function createPagination(currentPage, totalPages, onPageChange) {
  let html = '<div class="flex gap-2 justify-center items-center mt-6">';

  // Previous button
  html += `<button class="btn btn-secondary ${currentPage === 1 ? 'opacity-50 cursor-not-allowed' : ''}"
           onclick="${currentPage > 1 ? `${onPageChange}(${currentPage - 1})` : 'return false'}"
           ${currentPage === 1 ? 'disabled' : ''}>
           <i class="fas fa-chevron-left"></i>
           </button>`;

  // Page numbers
  const maxVisible = 5;
  let startPage = Math.max(1, currentPage - Math.floor(maxVisible / 2));
  let endPage = Math.min(totalPages, startPage + maxVisible - 1);

  if (endPage - startPage < maxVisible - 1) {
    startPage = Math.max(1, endPage - maxVisible + 1);
  }

  for (let i = startPage; i <= endPage; i++) {
    html += `<button class="btn ${i === currentPage ? 'btn-primary' : 'btn-secondary'}"
             onclick="${onPageChange}(${i})">${i}</button>`;
  }

  // Next button
  html += `<button class="btn btn-secondary ${currentPage === totalPages ? 'opacity-50 cursor-not-allowed' : ''}"
           onclick="${currentPage < totalPages ? `${onPageChange}(${currentPage + 1})` : 'return false'}"
           ${currentPage === totalPages ? 'disabled' : ''}>
           <i class="fas fa-chevron-right"></i>
           </button>`;

  html += '</div>';
  return html;
}

// CSS for animations
const style = document.createElement('style');
style.textContent = `
  @keyframes slideInRight {
    from {
      transform: translateX(100%);
      opacity: 0;
    }
    to {
      transform: translateX(0);
      opacity: 1;
    }
  }

  @keyframes slideOutRight {
    from {
      transform: translateX(0);
      opacity: 1;
    }
    to {
      transform: translateX(100%);
      opacity: 0;
    }
  }
`;
document.head.appendChild(style);
