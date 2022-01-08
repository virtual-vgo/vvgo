import { Link } from "react-router-dom";

const Admin = () => {
  return (
    <div>
      <h1>Executive Director Links</h1>
      <ul>
        <li>
          <Link to="/admin/mixtape/">View/edit mixtape projects.</Link>
        </li>
        <li>
          <Link to="/admin/sessions">View/edit login sessions.</Link>
        </li>
      </ul>
    </div>
  );
};

export default Admin;
