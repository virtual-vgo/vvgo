import _ from "lodash"
import Masonry from "@mui/lab/Masonry";
import Col from "react-bootstrap/Col";
import Row from "react-bootstrap/Row";
import {latestProject, Project, useCreditsTable, useProjects} from "../datasets";
import {AlertUnreleasedProject} from "./shared/AlertUnreleasedProject";
import {FancyProjectMenu, useMenuSelection} from "./shared/FancyProjectMenu";
import {LoadingText} from "./shared/LoadingText";
import {ProjectHeader} from "./shared/ProjectHeader";
import {RootContainer} from "./shared/RootContainer";
import {YoutubeIframe} from "./shared/YoutubeIframe";

const documentTitle = "Projects";
const permaLink = (project: Project) => `/projects/${project.Name}`;
const pathMatcher = /\/projects\/(.+)\/?/;

const searchProjects = (query: string, projects: Project[]): Project[] => {
    return _.defaultTo(projects, []).filter(r =>
        r.Name.toLowerCase().includes(query) ||
        r.Title.toLowerCase().includes(query) ||
        r.Sources.toLowerCase().includes(query));
};

export const Projects = () => {
    const allProjects = useProjects();
    const allowedProjects = _.defaultTo(allProjects, []).filter(r => !r.Hidden);
    const [selected, setSelected] = useMenuSelection(allowedProjects, pathMatcher, permaLink, latestProject(allowedProjects));

    if (!allProjects)
        return <RootContainer title={documentTitle}>
            <LoadingText/>
        </RootContainer>;

    return <RootContainer title={documentTitle}>
        <Row>
            <Col lg={3}>
                <FancyProjectMenu
                    choices={allowedProjects}
                    selected={selected}
                    setSelected={setSelected}
                    permaLink={permaLink}
                    searchChoices={searchProjects}
                    buttonContent={(proj: Project) =>
                        <div>
                            {proj.Title}
                            {!proj.PartsReleased ? <em><small><br/>Unreleased</small></em> : ""}
                            {!proj.VideoReleased ? <em><small><br/>In Production</small></em> : ""}
                            {proj.VideoReleased ? <em><small><br/>Completed</small></em> : ""}
                        </div>}/>
            </Col>
            <Col>
                {selected ?
                    <div className="mx-4">
                        <AlertUnreleasedProject project={selected}/>
                        <ProjectHeader project={selected}/>
                        {selected.PartsArchived ?
                            selected.YoutubeEmbed ?
                                <YoutubeIframe project={selected}/> :
                                <div className="text-center text-info">
                                    <em>Video coming soon!</em>
                                </div> :
                            <div/>}
                        <ProjectCredits project={selected}/>
                    </div> :
                    <div/>}
            </Col>
        </Row>
    </RootContainer>;
};

const ProjectCredits = (props: { project: Project }) => {
    const creditsTable = useCreditsTable(props.project);
    return <div>
        {_.defaultTo(creditsTable, []).map(topic => <Row key={topic.Name}>
            <Row>
                <Col className="text-center">
                    <h2><strong>— {topic.Name} —</strong></h2>
                </Col>
            </Row>
            <Row>
                <Masonry
                    columns={3}
                    spacing={1}
                    defaultHeight={450}
                    defaultColumns={3}
                    defaultSpacing={1}>
                    {_.isEmpty(topic.Rows) ? <div/> : topic.Rows.map(team =>
                        <Col key={team.Name} lg={4}>
                            <h5>{team.Name}</h5>
                            <ul className="list-unstyled">
                                {team.Rows.map(credit =>
                                    <li key={credit.Name}>{credit.Name} <small>{credit.BottomText}</small></li>)}
                            </ul>
                        </Col>)}
                </Masonry>
            </Row>
        </Row>)}
    </div>;
};


