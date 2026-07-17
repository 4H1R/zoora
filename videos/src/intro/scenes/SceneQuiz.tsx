import { AbsoluteFill, useCurrentFrame, useVideoConfig, interpolate } from 'remotion';
import { colors, font, radius, shadow, toFa } from '../../lib/tokens';
import { Chip, ToastMock } from '../../components/mock';
import { Icon } from '../../components/Icon';
import { Cursor } from '../../components/Cursor';
import { enter, pop, fadeUp } from '../../lib/anim';

const OPTIONS = ['نیوتن', 'ژول', 'وات', 'پاسکال'];

const OptionRow: React.FC<{ label: string; selected: boolean; selectProgress: number }> = ({
  label,
  selected,
  selectProgress,
}) => (
  <div
    dir="rtl"
    style={{
      display: 'flex',
      alignItems: 'center',
      gap: 14,
      padding: '16px 20px',
      borderRadius: radius.lg,
      border: `1.5px solid ${selected ? colors.green600 : colors.zinc200}`,
      background: selected ? colors.green50 : colors.white,
      fontFamily: font,
      fontSize: 19,
      fontWeight: 600,
      color: colors.zinc900,
    }}
  >
    <div
      style={{
        width: 22,
        height: 22,
        borderRadius: '50%',
        border: `2px solid ${selected ? colors.green600 : colors.zinc300}`,
        display: 'flex',
        alignItems: 'center',
        justifyContent: 'center',
        flexShrink: 0,
      }}
    >
      {selected ? (
        <div
          style={{
            width: 11 * selectProgress,
            height: 11 * selectProgress,
            borderRadius: '50%',
            background: colors.green600,
          }}
        />
      ) : null}
    </div>
    {label}
  </div>
);

export const SceneQuiz: React.FC = () => {
  const frame = useCurrentFrame();
  const { fps } = useVideoConfig();

  const cardIn = enter(frame, fps, 5);
  const remaining = 19 * 60 + 59 - Math.floor(frame / fps);
  const timer = `${String(Math.floor(remaining / 60)).padStart(2, '0')}:${String(remaining % 60).padStart(2, '0')}`;

  const selected = frame >= 152;
  const selectProgress = pop(frame, fps, 152);
  // Slide to next question
  const slide = interpolate(frame, [200, 222], [0, 1], {
    extrapolateLeft: 'clamp',
    extrapolateRight: 'clamp',
  });
  const questionNo = slide > 0.5 ? 4 : 3;
  const gradeIn = pop(frame, fps, 330);

  return (
    <AbsoluteFill
      style={{
        background: `linear-gradient(200deg, ${colors.zinc100}, #eef7f1)`,
        alignItems: 'center',
        justifyContent: 'center',
        fontFamily: font,
      }}
    >
      <div
        style={{
          width: 900,
          opacity: cardIn,
          transform: `translateY(${(1 - cardIn) * 50}px)`,
          display: 'flex',
          flexDirection: 'column',
          gap: 22,
          position: 'relative',
        }}
      >
        {/* Quiz header */}
        <div dir="rtl" style={{ display: 'flex', alignItems: 'center', justifyContent: 'space-between' }}>
          <div style={{ fontSize: 28, fontWeight: 800, color: colors.zinc900 }}>آزمون میان‌ترم فیزیک</div>
          <div
            style={{
              display: 'flex',
              alignItems: 'center',
              gap: 8,
              padding: '9px 18px',
              borderRadius: 999,
              background: colors.white,
              border: `1px solid ${colors.zinc200}`,
              boxShadow: shadow.md,
              color: colors.zinc800,
              fontSize: 18,
              fontWeight: 700,
            }}
          >
            <Icon name="clock" size={18} color={colors.warning} />
            {toFa(timer)}
          </div>
        </div>

        {/* Progress */}
        <div dir="rtl" style={{ display: 'flex', alignItems: 'center', gap: 14 }}>
          <span style={{ fontSize: 15, fontWeight: 600, color: colors.zinc500, whiteSpace: 'nowrap' }}>
            سؤال {toFa(questionNo)} از {toFa(20)}
          </span>
          <div style={{ flex: 1, height: 8, borderRadius: 999, background: colors.zinc200, overflow: 'hidden' }}>
            <div
              style={{
                width: `${(questionNo / 20) * 100}%`,
                height: '100%',
                background: colors.green600,
                borderRadius: 999,
              }}
            />
          </div>
        </div>

        {/* Question card (slides to next) */}
        <div style={{ position: 'relative', overflow: 'hidden', borderRadius: radius.xxl }}>
          <div
            dir="rtl"
            style={{
              background: colors.white,
              border: `1px solid ${colors.zinc200}`,
              borderRadius: radius.xxl,
              boxShadow: shadow.lg,
              padding: 30,
              display: 'flex',
              flexDirection: 'column',
              gap: 16,
              transform: `translateX(${slide * 110}%)`,
              opacity: 1 - slide,
            }}
          >
            <div style={{ fontSize: 22, fontWeight: 700, color: colors.zinc900, marginBottom: 6 }}>
              یکای اندازه‌گیری نیرو کدام است؟
            </div>
            {OPTIONS.map((o, i) => (
              <div key={o} style={fadeUp(frame, fps, 40 + i * 12, 20)}>
                <OptionRow label={o} selected={selected && i === 0} selectProgress={selectProgress} />
              </div>
            ))}
          </div>
          {slide > 0 ? (
            <div
              dir="rtl"
              style={{
                position: 'absolute',
                inset: 0,
                background: colors.white,
                border: `1px solid ${colors.zinc200}`,
                borderRadius: radius.xxl,
                boxShadow: shadow.lg,
                padding: 30,
                display: 'flex',
                flexDirection: 'column',
                gap: 16,
                transform: `translateX(${(slide - 1) * 110}%)`,
                opacity: slide,
              }}
            >
              <div style={{ fontSize: 22, fontWeight: 700, color: colors.zinc900, marginBottom: 6 }}>
                کدام رابطه، قانون دوم نیوتن را بیان می‌کند؟
              </div>
              {['F = ma', 'E = mc²', 'V = IR', 'P = F/A'].map((o) => (
                <OptionRow key={o} label={o} selected={false} selectProgress={0} />
              ))}
            </div>
          ) : null}
        </div>

        {/* Anti-cheat / feature chips */}
        <div dir="rtl" style={{ display: 'flex', gap: 14, justifyContent: 'center', marginTop: 8 }}>
          {['سؤال‌های درهم', 'نمرهٔ منفی', 'بانک سؤال', 'زمان‌بندی دقیق'].map((label, i) => (
            <div key={label} style={fadeUp(frame, fps, 240 + i * 12, 24)}>
              <Chip icon="check">{label}</Chip>
            </div>
          ))}
        </div>

        {/* Auto-grade toast */}
        {frame >= 330 ? (
          <div
            style={{
              position: 'absolute',
              top: -84,
              insetInlineStart: 0,
              transform: `scale(${gradeIn})`,
              transformOrigin: 'top right',
            }}
          >
            <ToastMock text={`نمره‌دهی خودکار انجام شد — ${toFa(18)} از ${toFa(20)}`} />
          </div>
        ) : null}

        <Cursor
          keys={[
            { frame: 90, x: 450, y: 560 },
            { frame: 140, x: 660, y: 262 },
            { frame: 152, x: 660, y: 262, click: true },
            { frame: 420, x: 660, y: 262 },
          ]}
        />
      </div>
    </AbsoluteFill>
  );
};
