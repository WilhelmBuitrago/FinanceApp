import React, { useState } from 'react';
import ImportStatus from '../components/ImportStatus';

// TODO: Integrate with backend API endpoints
// TODO: Implement file selection via Tauri dialog
// TODO: Poll job status

type JobStatus = 'UPLOAD' | 'PROCESSING' | 'VALIDATING' | 'IMPORTING' | 'SUCCESS' | 'ERROR';

export default function ImportPage() {
  const [file, setFile] = useState<File | null>(null);
  const [jobStatus, setJobStatus] = useState<JobStatus | null>(null);

  const handleFileSelect = () => {
    // TODO: Open Tauri file picker
    setFile(null);
  };

  const handleUpload = async () => {
    // TODO: POST /documents/upload
    setJobStatus('UPLOAD');
  };

  return (
    <div>
      <h1>Import Documents</h1>
      <button onClick={handleFileSelect}>Select File</button>
      {file && <p>Selected: {file.name}</p>}
      <button onClick={handleUpload} disabled={!file}>Upload</button>
      
      {jobStatus && (
        <div>
          <ImportStatus status={jobStatus} />
          <p>Current status: {jobStatus}</p>
        </div>
      )}
    </div>
  );
}
