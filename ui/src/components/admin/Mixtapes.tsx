import { isEmpty, uniq } from "lodash/fp";
import { CSSProperties, useRef, useState } from "react";
import { Dropdown, Table, Toast } from "react-bootstrap";
import ReactMarkdown from "react-markdown";
import { getSession } from "../../auth";
import { GuildMember } from "../../datasets";
import { Project } from "../../resources/mixtape/Project";
import { Resources, useResource } from "../../resources/Resources";
import { ChannelSchema } from "../../resources/schema/ChannelSchema";

const ManageMixtapes = () => {
  return (
    <div>
      <h1 className="mb-3">Mixtapes</h1>
      <ProjectTable />
    </div>
  );
};

export default ManageMixtapes;

const ProjectTable = () => {
  const db = new Resources(getSession().key);
  const [projects, setProjects] = useResource(db.mixtape.projects.list);
  const [guildMembers] = useResource(db.guildMembers.list);
  const [channels] = useResource(db.channels.list);
  return (
    <div>
      <Table variant="dark" bordered size="small">
        <CreateProjectRow projects={projects} setProjects={setProjects} />
        {projects
          ?.sort((a, b) => b.id - a.id)
          .map((proj) => [
            <thead key={`${proj.id}-head`}>
              <tr>
                <td className="text-muted">id</td>
                <th>Title</th>
                <th>Name</th>
                <th>Mixtape</th>
                <th>Channel</th>
                <th>Hosts</th>
                <th />
              </tr>
            </thead>,
            <ProjectRow
              key={`${proj.id}-body`}
              projects={projects}
              setProjects={setProjects}
              thisProject={proj}
              channels={channels ?? []}
              guildMembers={guildMembers ?? []}
            />,
          ])}
      </Table>
    </div>
  );
};

const CreateProjectRow = (props: {
  projects: Project[] | undefined;
  setProjects: (val: Project[]) => void;
}) => {
  const db = new Resources(getSession().key).mixtape.projects;
  const createProject = () => {
    db.create().then((proj) => {
      props.setProjects([proj, ...(props.projects ?? [])]);
    });
  };
  return (
    <>
      <thead>
        <tr>
          <td className="text-muted">id</td>
          <th>Title</th>
          <th>Name</th>
          <th>Mixtape</th>
          <th>Channel</th>
          <th>Hosts</th>
          <th />
        </tr>
      </thead>
      <tbody>
        <tr>
          <td
            colSpan={7}
            className="text-underline text-primary"
            style={{ cursor: "pointer" }}
            onClick={createProject}
          >
            + new project
          </td>
        </tr>
      </tbody>
    </>
  );
};

const ProjectRow = (props: {
  thisProject: Project | undefined;
  projects: Project[] | undefined;
  setProjects: (val: Project[]) => void;
  guildMembers: GuildMember[];
  channels: ChannelSchema[];
}) => {
  const [project, setProject] = useState(props.thisProject);
  const db = new Resources(getSession().key).mixtape.projects;

  const deleteProject = () => {
    const id = project?.id ?? 0;
    if (id == 0) return;
    db.delete(id).then(() => {
      setProject(undefined);
      props.setProjects(props.projects?.filter((p) => p.id != id) ?? []);
    });
  };
  if (!project) return <div />;
  return (
    <tbody>
      <tr>
        <td className="text-muted">
          <em># {project.id}</em>
        </td>
        <EditField
          project={project}
          setProject={setProject}
          initValue={project.title}
          setField={(val, proj) => (proj.title = val)}
        />
        <EditField
          project={project}
          setProject={setProject}
          initValue={project.Name}
          setField={(val, proj) => (proj.Name = val)}
        />
        <EditField
          project={project}
          setProject={setProject}
          initValue={project.mixtape}
          setField={(val, proj) => (proj.mixtape = val)}
        />
        <EditChannel
          project={project}
          setProject={setProject}
          channels={props.channels}
        />
        <EditHosts
          project={project}
          setProject={setProject}
          guildMembers={props.guildMembers}
        />
        <td>
          <span
            onClick={() => deleteProject()}
            style={{ cursor: "pointer" }}
            className="text-primary"
          >
            delete
          </span>
        </td>
      </tr>
      <tr>
        <td colSpan={7}>
          <ReactMarkdown>{props.thisProject?.blurb ?? ""}</ReactMarkdown>
        </td>
      </tr>
    </tbody>
  );
};

const EditField = (props: {
  project: Project;
  setProject: (proj: Project) => void;
  initValue: string;
  setField: (val: string, proj: Project) => void;
}) => {
  const [tdClassName, setTdClassName] = useState("");
  const inputStyle: CSSProperties = {
    height: "100%",
    width: "100%",
  };
  const inputRef = useRef<HTMLInputElement>(null);

  const db = new Resources(getSession().key).mixtape.projects;

  const saveProject = () => {
    const newVal = inputRef.current?.value ?? props.initValue;
    if (props.project.id != 0 && props.initValue == newVal) return;

    const project = props.project;
    props.setField(newVal, project);
    (project.id == 0 ? db.create(project) : db.save(project))
      .then((proj) => {
        setTdClassName("");
        props.setProject(proj);
        console.log("saved project", proj);
      })
      .catch(() => setTdClassName("text-warning"));
  };

  return (
    <td
      className={tdClassName}
      onKeyUp={(e) => {
        if (["Enter", "Escape"].includes(e.code)) saveProject();
      }}
      onBlur={() => saveProject()}
    >
      <input
        style={inputStyle}
        ref={inputRef}
        type="text"
        defaultValue={props.initValue}
      />
    </td>
  );
};

const EditChannel = (props: {
  project: Project;
  setProject: (proj: Project) => void;
  channels: ChannelSchema[];
}) => {
  return (
    <td>
      <select>
        <option>
          {props.project?.channel ? props.project.channel : "choose a channel"}
        </option>
        {props.channels?.map((r, i) => (
          <option key={i}>r.name</option>
        ))}
      </select>
    </td>
  );
};

const EditHosts = (props: {
  project: Project;
  setProject: (proj: Project | undefined) => void;
  guildMembers: GuildMember[];
}) => {
  const [showToast, setShowToast] = useState(false);
  const [searchQuery, setSearchQuery] = useState("");

  const db = new Resources(getSession().key).mixtape.projects;

  const addHost = (member: GuildMember | undefined) => {
    if (!member) return;
    const project = props.project;
    project.hosts = uniq([...project.hosts, member.user.id]);
    db.save(project).then((result) => {
      props.setProject(result);
    });
  };

  const rmHost = (member: GuildMember | undefined) => {
    if (!member) return;
    const project = props.project;
    project.hosts = project.hosts.filter((id) => id != member.user.id);
    db.save(project).then((result) => props.setProject(result));
  };

  const projectMembers = props.guildMembers.filter((m) =>
    props.project.hosts.includes(m.user.id)
  );

  const filteredMembers =
    isEmpty(props.guildMembers) || isEmpty(searchQuery) || !showToast
      ? []
      : props.guildMembers
          ?.filter((m) => !props.project.hosts.includes(m.user.id))
          .filter(
            (m) =>
              m.user.username.toLowerCase().includes(searchQuery) ||
              m.nick.toLowerCase().includes(searchQuery) ||
              m.user.id.toString().includes(searchQuery)
          );

  const hostList = projectMembers.map((m, i) => (
    <li key={i}>
      {m.displayName()}{" "}
      <span
        className="text-primary text-decoration-underline"
        style={{ cursor: "pointer" }}
        onClick={() => rmHost(m)}
      >
        remove
      </span>
    </li>
  ));

  return (
    <td>
      <ul>
        {isEmpty(hostList) ? <li>This project has no hosts.</li> : hostList}
      </ul>
      <input
        placeholder="search nicks"
        defaultValue={searchQuery}
        onChange={(e) => {
          setShowToast(true);
          setSearchQuery(e.target.value?.toLowerCase() ?? "");
        }}
        onKeyUp={(e) => ["Escape"].includes(e.code) && setShowToast(false)}
      />
      <div>
        <Toast>
          {filteredMembers
            ?.filter((m) => !isEmpty(m.nick))
            .filter((m) => !props.project.hosts.includes(m.user.id))
            .slice(0, 5)
            .map((m, i) => (
              <Dropdown.Item key={i} onClick={() => addHost(m)}>
                {m.nick}
              </Dropdown.Item>
            ))}
        </Toast>
      </div>
    </td>
  );
};
