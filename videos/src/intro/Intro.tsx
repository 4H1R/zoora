import { AbsoluteFill } from 'remotion';
import { TransitionSeries, linearTiming } from '@remotion/transitions';
import { fade } from '@remotion/transitions/fade';
import { Subtitles } from '../lib/Subtitles';
import { LINES, D, T, INTRO_DURATION, FPS } from './script';
import { SceneLogo } from './scenes/SceneLogo';
import { SceneClasses } from './scenes/SceneClasses';
import { SceneLive } from './scenes/SceneLive';
import { SceneQuiz } from './scenes/SceneQuiz';
import { SceneReach } from './scenes/SceneReach';
import { SceneOutro } from './scenes/SceneOutro';

export { INTRO_DURATION, FPS };

const transition = () => (
  <TransitionSeries.Transition presentation={fade()} timing={linearTiming({ durationInFrames: T })} />
);

export const Intro: React.FC = () => {
  return (
    <AbsoluteFill>
      <TransitionSeries>
        <TransitionSeries.Sequence durationInFrames={D.logo}>
          <SceneLogo />
        </TransitionSeries.Sequence>
        {transition()}
        <TransitionSeries.Sequence durationInFrames={D.classes}>
          <SceneClasses />
        </TransitionSeries.Sequence>
        {transition()}
        <TransitionSeries.Sequence durationInFrames={D.live}>
          <SceneLive />
        </TransitionSeries.Sequence>
        {transition()}
        <TransitionSeries.Sequence durationInFrames={D.quiz}>
          <SceneQuiz />
        </TransitionSeries.Sequence>
        {transition()}
        <TransitionSeries.Sequence durationInFrames={D.reach}>
          <SceneReach />
        </TransitionSeries.Sequence>
        {transition()}
        <TransitionSeries.Sequence durationInFrames={D.outro}>
          <SceneOutro />
        </TransitionSeries.Sequence>
      </TransitionSeries>
      <Subtitles lines={LINES} />
    </AbsoluteFill>
  );
};
