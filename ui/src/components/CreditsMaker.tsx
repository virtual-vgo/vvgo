import {CSSProperties, useRef, useState} from "react";
import {Dropdown} from "react-bootstrap";
import Button from "react-bootstrap/Button";
import Col from "react-bootstrap/Col";
import FormControl from "react-bootstrap/FormControl";
import FormLabel from "react-bootstrap/FormLabel";
import Row from "react-bootstrap/Row";
import {links} from "../data/links";
import {CreditsPasta, fetchApi, Project, useProjects} from "../datasets";

const DefaultSpreadsheetId = "1a-2u726Hg-Wp5GMWfLnYwSi2DvTMym85gQqpRviafJk";
const DefaultReadRange = "06 Aurene!A3:I39";

export const CreditsMaker = () => {
    const [pasta, setPasta] = useState({} as CreditsPasta);
    return <div>
        <Row>
            <Col><InputForm setPasta={setPasta}/></Col>
        </Row>
        <Row>
            <Col className="mt-4"><h2 className="text-center">~ Results ~</h2></Col>
        </Row>
        <Row>
            <Col><PastaResult title={"Website Pasta"} content={pasta.WebsitePasta}/></Col>
            <Col><PastaResult title={"Video Pasta"} content={pasta.VideoPasta}/></Col>
            <Col><PastaResult title={"Youtube Pasta"} content={pasta.YoutubePasta}/></Col>
        </Row>
    </div>;
};

const InputForm = (props: { setPasta: (pasta: CreditsPasta) => void; }) => {
    const [errorMessage, setErrorMessage] = useState("");
    const [projectName, setProjectName] = useState("");
    const spreadsheetIDInputRef = useRef({} as HTMLInputElement);
    const readRangeInputRef = useRef({} as HTMLInputElement);

    const onClickSubmit = () => {
        const params = new URLSearchParams({
            spreadsheetID: spreadsheetIDInputRef.current.value,
            readRange: readRangeInputRef.current.value,
            projectName: projectName,
        });

        fetchApi("/credits/pasta?" + params.toString(), {method: "GET"})
            .then(resp => {
                setErrorMessage("");
                props.setPasta(resp.CreditsPasta ?? {WebsitePasta: "", VideoPasta: "", YoutubePasta: ""});
            })
            .catch(err => setErrorMessage(err.toString()));
    };

    const helpLink = (to: string) =>
        <small className="form-text text-muted">
            <a href={to}>Help</a>
        </small>;

    return <div>
        <Row>
            <Col className="m-2">
                <FormLabel>
                    Spreadsheet ID {helpLink(links.Help.SpreadsheetId)}
                </FormLabel>
                <FormControl
                    type="text"
                    defaultValue={DefaultSpreadsheetId}
                    ref={spreadsheetIDInputRef}/>
            </Col>

            <Col className="m-2">
                <FormLabel>
                    Read Range {helpLink(links.Help.ReadRange)}
                </FormLabel>
                <FormControl
                    type="text"
                    defaultValue={DefaultReadRange}
                    ref={readRangeInputRef}/>
            </Col>
            <Col className="m-2">
                <FormLabel>
                    Project Name
                </FormLabel>
                <ProjectDropdown
                    projectName={projectName}
                    setProjectName={setProjectName}/>
            </Col>
        </Row>

        <Row>
            {errorMessage != "" ?
                <Col className="text-warning mx-2"><code>{errorMessage}</code></Col> :
                <div/>}
        </Row>

        <Row>
            <Col className="m-2">
                <Button variant="primary" onClick={onClickSubmit}>Submit</Button>
            </Col>
        </Row>
    </div>;
};

const ProjectDropdown = (props: { projectName: string, setProjectName: (name: string) => void }) => {
    const projects = useProjects() ?? [];
    const [buttonText, setButtonText] = useState("select project");

    const updateSelect = (project: Project) => {
        props.setProjectName(project.Name);
        setButtonText(project.Title);
    };

    return <div className="d-flex">
        <Dropdown>
            <Dropdown.Toggle variant="light">
                {buttonText}
            </Dropdown.Toggle>

            <Dropdown.Menu>
                {projects.map(p => <Dropdown.Item key={p.Name} onClick={() => updateSelect(p)}>
                    {p.Title}
                </Dropdown.Item>)}
            </Dropdown.Menu>
        </Dropdown>
    </div>;
};

const PastaResult = (props: { title: string, content: string }) => {
    const textAreaStyle = ({
        whiteSpace: "pre",
        overflowWrap: "normal",
        overflowX: "scroll",
    } as CSSProperties);

    return <Row className="row-cols-1">
        <Col className="mt-4"><h3 className="text-center">{props.title}</h3></Col>
        <Col>
            <textarea
                readOnly
                className="text-monospace form-control"
                rows={10}
                style={textAreaStyle}
                value={props.content}/>
        </Col>
        <Col className="mt-2">
            <Button
                variant="secondary"
                onClick={() => navigator.clipboard.writeText(props.content)}>
                Copy text
            </Button>
        </Col>
    </Row>;
};

export default CreditsMaker;
