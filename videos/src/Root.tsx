import '@fontsource-variable/vazirmatn';
import { Composition } from 'remotion';
import { Intro, INTRO_DURATION, FPS } from './intro/Intro';

export const Root: React.FC = () => {
  return (
    <>
      <Composition
        id="Intro"
        component={Intro}
        durationInFrames={INTRO_DURATION}
        fps={FPS}
        width={1920}
        height={1080}
      />
    </>
  );
};
