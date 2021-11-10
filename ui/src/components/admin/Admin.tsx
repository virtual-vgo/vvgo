import { Link } from "react-router-dom";

const Admin = () => {
  return (
    <div>
      <h1>Executive Director Links</h1>
      <ul>
        <li>
          <Link to="/admin/mixtape/">Manage Mixtape Projects</Link>
        </li>
        <li>
          <Link to="/mixtape/NewProjectWorkflow">
            Mixtape New Project Workflow
          </Link>
        </li>
      </ul>
    </div>
  );
};

export default Admin;
