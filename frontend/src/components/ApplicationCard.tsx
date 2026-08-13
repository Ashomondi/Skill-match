import React, { useState } from 'react';
import { Application, UpdateApplicationDTO } from '../../services/applications';

interface ApplicationCardProps {
  application: Application;
  onUpdate: (id: string, data: UpdateApplicationDTO) => Promise<void>;
  onDelete: (id: string) => Promise<void>;
}

const statusColors: Record<Application['status'], string> = {
  Applied: 'bg-blue-100 text-blue-800',
  'Phone Screening': 'bg-yellow-100 text-yellow-800',
  Interviewing: 'bg-purple-100 text-purple-800',
  Offered: 'bg-green-100 text-green-800',
  Rejected: 'bg-red-100 text-red-800',
  Withdrawn: 'bg-gray-100 text-gray-800',
};

export const ApplicationCard: React.FC<ApplicationCardProps> = ({
  application,
  onUpdate,
  onDelete,
}) => {
  const [isEditing, setIsEditing] = useState(false);
  const [status, setStatus] = useState<Application['status']>(application.status);
  const [notes, setNotes] = useState(application.notes || '');
  const [loading, setLoading] = useState(false);

  const handleSave = async () => {
    try {
      setLoading(true);
      await onUpdate(application.id, { status, notes });
      setIsEditing(false);
    } catch (err) {
      console.error('Failed to update application', err);
    } finally {
      setLoading(false);
    }
  };

  const handleDelete = async () => {
    if (window.confirm('Are you sure you want to delete this application record?')) {
      try {
        setLoading(true);
        await onDelete(application.id);
      } catch (err) {
        console.error('Failed to delete application', err);
        setLoading(false);
      }
    }
  };

  return (
    <div className="bg-white rounded-lg shadow-md p-6 border border-gray-200 flex flex-col justify-between">
      <div>
        <div className="flex justify-between items-start">
          <div>
            <h3 className="text-lg font-bold text-gray-900">{application.jobTitle}</h3>
            <p className="text-md font-medium text-gray-600">{application.company}</p>
            {application.location && (
              <p className="text-sm text-gray-500 mt-1">{application.location}</p>
            )}
          </div>
          <span
            className={`px-3 py-1 text-xs font-semibold rounded-full ${
              statusColors[application.status]
            }`}
          >
            {application.status}
          </span>
        </div>

        <div className="mt-4 text-sm text-gray-500">
          <p>Applied on: {new Date(application.appliedDate).toLocaleDateString()}</p>
        </div>

        {isEditing ? (
          <div className="mt-4 space-y-3">
            <div>
              <label className="block text-xs font-medium text-gray-700">Status</label>
              <select
                value={status}
                onChange={(e) => setStatus(e.target.value as Application['status'])}
                className="mt-1 block w-full rounded-md border-gray-300 shadow-sm focus:border-indigo-500 focus:ring-indigo-500 sm:text-sm border p-2"
              >
                <option value="Applied">Applied</option>
                <option value="Phone Screening">Phone Screening</option>
                <option value="Interviewing">Interviewing</option>
                <option value="Offered">Offered</option>
                <option value="Rejected">Rejected</option>
                <option value="Withdrawn">Withdrawn</option>
              </select>
            </div>
            <div>
              <label className="block text-xs font-medium text-gray-700">Notes</label>
              <textarea
                value={notes}
                onChange={(e) => setNotes(e.target.value)}
                rows={2}
                className="mt-1 block w-full rounded-md border-gray-300 shadow-sm focus:border-indigo-500 focus:ring-indigo-500 sm:text-sm border p-2"
                placeholder="Add notes about your interview or recruiter..."
              />
            </div>
          </div>
        ) : (
          application.notes && (
            <div className="mt-4 bg-gray-50 p-3 rounded-md text-sm text-gray-700">
              <span className="font-semibold block text-xs text-gray-500 mb-1">Notes:</span>
              {application.notes}
            </div>
          )
        )}
      </div>

      <div className="mt-6 flex justify-end space-x-2 pt-4 border-t border-gray-100">
        {isEditing ? (
          <>
            <button
              onClick={() => setIsEditing(false)}
              disabled={loading}
              className="px-3 py-1.5 border border-gray-300 text-xs font-medium rounded text-gray-700 bg-white hover:bg-gray-50"
            >
              Cancel
            </button>
            <button
              onClick={handleSave}
              disabled={loading}
              className="px-3 py-1.5 border border-transparent text-xs font-medium rounded text-white bg-indigo-600 hover:bg-indigo-700"
            >
              {loading ? 'Saving...' : 'Save'}
            </button>
          </>
        ) : (
          <>
            <button
              onClick={() => setIsEditing(true)}
              className="px-3 py-1.5 border border-gray-300 text-xs font-medium rounded text-gray-700 bg-white hover:bg-gray-50"
            >
              Edit
            </button>
            <button
              onClick={handleDelete}
              disabled={loading}
              className="px-3 py-1.5 border border-transparent text-xs font-medium rounded text-white bg-red-600 hover:bg-red-700"
            >
              Delete
            </button>
          </>
        )}
      </div>
    </div>
  );
};
