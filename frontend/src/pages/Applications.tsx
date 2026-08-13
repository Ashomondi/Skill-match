import React, { useEffect, useState } from 'react';
import { applicationService, Application, UpdateApplicationDTO } from '../services/applications';
import { ApplicationCard } from '../components/applications/ApplicationCard';

export const Applications: React.FC = () => {
  const [applications, setApplications] = useState<Application[]>([]);
  const [loading, setLoading] = useState<boolean>(true);
  const [error, setError] = useState<string | null>(null);

  const fetchApplications = async () => {
    try {
      setLoading(true);
      setError(null);
      const data = await applicationService.getApplications();
      setApplications(data);
    } catch (err: any) {
      setError(err.message || 'Failed to load applications');
    } finally {
      setLoading(false);
    }
  };

  useEffect(() => {
    fetchApplications();
  }, []);

  const handleUpdate = async (id: string, data: UpdateApplicationDTO) => {
    const updated = await applicationService.updateApplication(id, data);
    setApplications((prev) =>
      prev.map((app) => (app.id === id ? updated : app))
    );
  };

  const handleDelete = async (id: string) => {
    await applicationService.deleteApplication(id);
    setApplications((prev) => prev.filter((app) => app.id !== id));
  };

  if (loading) {
    return (
      <div className="flex justify-center items-center h-96">
        <div className="animate-spin rounded-full h-12 w-12 border-b-2 border-indigo-600"></div>
      </div>
    );
  }

  if (error) {
    return (
      <div className="max-w-7xl mx-auto px-4 py-8">
        <div className="bg-red-50 border-l-4 border-red-400 p-4">
          <div className="flex">
            <div className="ml-3">
              <p className="text-sm text-red-700">{error}</p>
              <button
                onClick={fetchApplications}
                className="mt-2 text-sm font-medium text-red-700 underline hover:text-red-600"
              >
                Try again
              </button>
            </div>
          </div>
        </div>
      </div>
    );
  }

  return (
    <div className="max-w-7xl mx-auto px-4 sm:px-6 lg:px-8 py-8">
      <div className="flex justify-between items-center mb-6">
        <div>
          <h1 className="text-2xl font-bold text-gray-900">Job Applications</h1>
          <p className="text-sm text-gray-500 mt-1">
            Track and manage your ongoing job hunt pipeline.
          </p>
        </div>
        <span className="bg-indigo-100 text-indigo-800 text-sm font-semibold px-3 py-1 rounded-full">
          Total: {applications.length}
        </span>
      </div>

      {applications.length === 0 ? (
        <div className="text-center bg-white rounded-lg shadow-sm border border-gray-200 p-12">
          <svg
            className="mx-auto h-12 w-12 text-gray-400"
            fill="none"
            viewBox="0 0 24 24"
            stroke="currentColor"
            aria-hidden="true"
          >
            <path
              strokeLinecap="round"
              strokeLinejoin="round"
              strokeWidth="2"
              d="M9 12h6m-6 4h6m2 5H7a2 2 0 01-2-2V5a2 2 0 012-2h5.586a1 1 0 01.707.293l5.414 5.414a1 1 0 01.293.707V19a2 2 0 01-2 2z"
            />
          </svg>
          <h3 className="mt-2 text-sm font-medium text-gray-900">No applications tracked</h3>
          <p className="mt-1 text-sm text-gray-500">
            Get started by exploring recommendations and saving or tracking applications.
          </p>
        </div>
      ) : (
        <div className="grid grid-cols-1 gap-6 sm:grid-cols-2 lg:grid-cols-3">
          {applications.map((app) => (
            <ApplicationCard
              key={app.id}
              application={app}
              onUpdate={handleUpdate}
              onDelete={handleDelete}
            />
          ))}
        </div>
      )}
    </div>
  );
};
