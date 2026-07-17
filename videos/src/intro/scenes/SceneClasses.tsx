import { AbsoluteFill, useCurrentFrame, useVideoConfig, interpolate } from 'remotion';
import { colors, font, radius, shadow } from '../../lib/tokens';
import { BrowserFrame, SidebarMock, ButtonMock, ClassCard, ToastMock } from '../../components/mock';
import { Cursor } from '../../components/Cursor';
import { enter, pop, fadeUp } from '../../lib/anim';

const NAV = [
  { icon: 'grid' as const, label: 'نمای کلی' },
  { icon: 'book' as const, label: 'کلاس‌ها', active: true },
  { icon: 'calendar' as const, label: 'برنامه زمانی' },
  { icon: 'video' as const, label: 'ضبط‌ها' },
  { icon: 'users' as const, label: 'دانشجویان' },
  { icon: 'chart' as const, label: 'تحلیل‌ها' },
  { icon: 'gear' as const, label: 'تنظیمات' },
];

const Field: React.FC<{ label: string; value: string; progress: number }> = ({
  label,
  value,
  progress,
}) => (
  <div dir="rtl" style={{ display: 'flex', flexDirection: 'column', gap: 8 }}>
    <span style={{ fontSize: 14, fontWeight: 600, color: colors.zinc700 }}>{label}</span>
    <div
      style={{
        height: 44,
        borderRadius: radius.lg,
        border: `1px solid ${colors.zinc200}`,
        background: colors.white,
        display: 'flex',
        alignItems: 'center',
        paddingInline: 14,
        fontSize: 15,
        color: colors.zinc900,
        overflow: 'hidden',
      }}
    >
      <span style={{ whiteSpace: 'nowrap' }}>
        {value.slice(0, Math.round(value.length * Math.min(1, Math.max(0, progress))))}
      </span>
      {progress > 0 && progress < 1 ? (
        <span style={{ width: 2, height: 20, background: colors.green600, marginInlineStart: 2 }} />
      ) : null}
    </div>
  </div>
);

export const SceneClasses: React.FC = () => {
  const frame = useCurrentFrame();
  const { fps } = useVideoConfig();

  const frameIn = enter(frame, fps, 5);
  const modalIn = pop(frame, fps, 235);
  const modalOut = interpolate(frame, [340, 352], [1, 0], {
    extrapolateLeft: 'clamp',
    extrapolateRight: 'clamp',
  });
  const showModal = frame >= 235 && frame < 352;
  const toastIn = pop(frame, fps, 356);
  const newCardIn = pop(frame, fps, 360);

  return (
    <AbsoluteFill
      style={{
        background: `linear-gradient(160deg, ${colors.zinc100}, #e7f6ec)`,
        alignItems: 'center',
        justifyContent: 'center',
      }}
    >
      <div
        style={{
          opacity: frameIn,
          transform: `translateY(${(1 - frameIn) * 60}px) scale(${0.96 + frameIn * 0.04})`,
        }}
      >
        <BrowserFrame>
          <div dir="rtl" style={{ display: 'flex', height: '100%', background: colors.white }}>
            <SidebarMock items={NAV} />
            <div style={{ flex: 1, padding: '30px 36px', position: 'relative' }}>
              {/* Page header */}
              <div
                style={{
                  display: 'flex',
                  alignItems: 'center',
                  justifyContent: 'space-between',
                  marginBottom: 30,
                }}
              >
                <div style={{ fontFamily: font, fontSize: 27, fontWeight: 800, color: colors.zinc900 }}>
                  کلاس‌ها
                </div>
                <ButtonMock icon="plus">کلاس جدید</ButtonMock>
              </div>

              {/* Class cards */}
              <div style={{ display: 'flex', gap: 24, flexWrap: 'wrap' }}>
                <div style={fadeUp(frame, fps, 60)}>
                  <ClassCard title="ریاضی دهم" students={31} next="فردا ۱۶:۰۰" hue={1} />
                </div>
                <div style={fadeUp(frame, fps, 78)}>
                  <ClassCard title="شیمی آلی" students={18} next="سه‌شنبه ۱۰:۰۰" hue={3} />
                </div>
                {frame >= 360 ? (
                  <div style={{ transform: `scale(${newCardIn})`, transformOrigin: 'center' }}>
                    <ClassCard title="فیزیک پیشرفته" students={24} next="امروز ۱۸:۰۰" hue={0} />
                  </div>
                ) : null}
              </div>

              {/* Toast */}
              {frame >= 356 ? (
                <div
                  style={{
                    position: 'absolute',
                    bottom: 26,
                    insetInlineStart: 30,
                    transform: `translateY(${(1 - toastIn) * 40}px)`,
                    opacity: toastIn,
                  }}
                >
                  <ToastMock text="کلاس «فیزیک پیشرفته» با موفقیت ساخته شد" />
                </div>
              ) : null}
            </div>
          </div>

          {/* Create-class modal */}
          {showModal ? (
            <AbsoluteFill
              style={{
                background: `rgba(9,9,11,${0.35 * Math.min(modalIn, modalOut)})`,
                alignItems: 'center',
                justifyContent: 'center',
              }}
            >
              <div
                dir="rtl"
                style={{
                  width: 520,
                  borderRadius: radius.xxl,
                  background: colors.white,
                  boxShadow: shadow.xl,
                  padding: 30,
                  fontFamily: font,
                  display: 'flex',
                  flexDirection: 'column',
                  gap: 20,
                  transform: `scale(${0.9 + 0.1 * Math.min(modalIn, modalOut)})`,
                  opacity: Math.min(modalIn, modalOut),
                }}
              >
                <div style={{ fontSize: 21, fontWeight: 800, color: colors.zinc900 }}>ساخت کلاس جدید</div>
                <Field label="نام کلاس" value="فیزیک پیشرفته" progress={(frame - 255) / 40} />
                <Field label="توضیحات" value="آمادگی کنکور — پایهٔ دوازدهم" progress={(frame - 285) / 35} />
                <div style={{ display: 'flex', justifyContent: 'flex-start', marginTop: 6 }}>
                  <ButtonMock>ایجاد کلاس</ButtonMock>
                </div>
              </div>
            </AbsoluteFill>
          ) : null}

          {/* Cursor above everything inside the browser frame */}
          <Cursor
            keys={[
              { frame: 140, x: 760, y: 520 },
              { frame: 200, x: 205, y: 68 },
              { frame: 212, x: 205, y: 68, click: true },
              { frame: 250, x: 205, y: 68 },
              { frame: 300, x: 880, y: 580 },
              { frame: 330, x: 948, y: 528, click: true },
              { frame: 420, x: 948, y: 528 },
            ]}
          />
        </BrowserFrame>
      </div>
    </AbsoluteFill>
  );
};
