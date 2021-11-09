import { Project } from "../../datasets";

export const AlertUnreleasedProject = (props: { project: Project }) => {
  if (props.project.PartsReleased) return <div />;
  return (
    <div className="alert alert-warning">
      This project is unreleased and invisible to members!
    </div>
  );
};
