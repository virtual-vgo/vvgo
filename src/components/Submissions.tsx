import { CSSProperties, useCallback, useMemo, useState } from "react";

import { Button } from "react-bootstrap";
import { Instrument } from "../datasets/Instrument";
import { useInstruments } from "../datasets";

const styles = {
  Form: {
    width: "100%",
    maxWidth: "500px",
    padding: "15px",
    margin: "auto",
  } as CSSProperties,
};

interface SubmissionResults {
  projectName: string;
  creditedName: string;
  partName: Instrument;
  instrumentPlayed: string;
  videoFiles: File[];
  audioFiles: File[];
}

export const Submissions = () => {
  const [differentInstrument, setDifferentInstrument] = useState(false);
  const [hasVideo, setHasVideo] = useState(false);

  const instruments = useInstruments()?.filter((i) => {
    var index = Number(i.instrumentIndex);
    return index !== 0 && index < 800;
  });

  const projectName = useMemo(() => {
    const href = window.location.href || "";
    return href.substring(href.lastIndexOf("/") + 1);
  }, []);

  const onFormSubmit = useCallback(
    (event: React.FormEvent<HTMLFormElement>) => {
      event.preventDefault();
      const target = event.target as any;
      const submissionResults: SubmissionResults = {
        projectName: target.projectName?.value,
        creditedName: target.creditedName?.value,
        partName:
          target.partName && (JSON.parse(target.partName?.value) as Instrument),
        instrumentPlayed: target.instrumentPlayed?.value,
        videoFiles: target.videoFiles?.files,
        audioFiles: target.audioFiles?.files,
      };

      const fileName = `${
        submissionResults.partName.partID
      }_${submissionResults.partName.partName
        .toLowerCase()
        .replace(/ /g, "")}_${encodeURI(submissionResults.creditedName)}_(${(
        submissionResults.instrumentPlayed ||
        submissionResults.partName.partNameStripped
      ).toLowerCase()})`;

      console.dir(submissionResults);
      console.log(fileName);
    },
    []
  );

  return (
    <div>
      <form className="mx-auto" style={styles.Form} onSubmit={onFormSubmit}>
        <div className="form-group">
          <h1>Project Submissions</h1>
          <label htmlFor="projectName">Project Name</label>
          <input
            className="form-control mb-1"
            type="text"
            id="projectName"
            name="projectName"
            value={projectName}
            disabled={true}
          />

          <label htmlFor="creditedName">Credited Name</label>
          <input
            className="form-control mb-1"
            type="text"
            id="creditedName"
            name="creditedName"
            placeholder="Your name as it will appear in credits"
          />

          <label htmlFor="partName">Part Name (on sheet music)</label>
          <select className="form-control mb-1" id="partName" name="partName">
            {instruments
              ?.map((i) => {
                if (!i.partName) {
                  return;
                }

                return (
                  <option key={i.partName} value={JSON.stringify(i)}>
                    {i.partName}
                  </option>
                );
              })
              .sort((a, b) => {
                var aKey: string = (a?.key || "").toString();
                var bKey: string = (b?.key || "").toString();
                return aKey.localeCompare(bKey);
              })}
          </select>

          <label htmlFor="differentInstrument">
            Did you play a different instrument?
          </label>
          <input
            className="form-check"
            type="checkbox"
            id="differentInstrument"
            name="differentInstrument"
            onChange={(e) => setDifferentInstrument(e.target.checked)}
          />

          {differentInstrument && (
            <>
              <label htmlFor="instrumentPlayed">
                Instrument Played (if different)
              </label>
              <select
                className="form-control mb-1"
                id="playedInstrument"
                name="playedInstrument"
              >
                {instruments
                  ?.filter(
                    (v, i, a) =>
                      a.findIndex(
                        (v2) => v2.partNameStripped === v.partNameStripped
                      ) === i
                  )
                  ?.map((i) => {
                    if (!i.partName) {
                      return;
                    }

                    return (
                      <option
                        key={i.partNameStripped}
                        value={i.partNameStripped}
                      >
                        {i.partNameStripped}
                      </option>
                    );
                  })
                  .sort((a, b) => {
                    var aKey: string = (a?.key || "").toString();
                    var bKey: string = (b?.key || "").toString();
                    return aKey.localeCompare(bKey);
                  })}
              </select>
            </>
          )}

          <label htmlFor="hasVideo">Has Video</label>
          <input
            className="form-check"
            type="checkbox"
            id="hasVideo"
            name="hasVideo"
            onChange={(e) => setHasVideo(e.target.checked)}
          />

          {hasVideo && (
            <>
              <label htmlFor="videoFiles">Video Files</label>
              <input
                className="form-control mb-1"
                type="file"
                id="videoFiles"
                name="videoFiles"
                multiple={true}
              />
            </>
          )}

          <label htmlFor="audioFiles">Audio Files</label>
          <input
            className="form-control mb-1"
            type="file"
            id="audioFiles"
            name="audioFiles"
            multiple={true}
          />
        </div>
        <div className="d-grid">
          <Button type="submit" variant={"primary"}>
            Submit
          </Button>
        </div>
      </form>
    </div>
  );
};
export default Submissions;
