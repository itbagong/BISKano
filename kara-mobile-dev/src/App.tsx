/**
 * Sample React Native App
 * https://github.com/facebook/react-native
 *
 * Generated with the UI Kitten TypeScript template
 * https://github.com/akveo/react-native-ui-kitten
 *
 * Documentation: https://akveo.github.io/react-native-ui-kitten/docs
 *
 * @format
 */
import 'react-native-reanimated';
import 'react-native-gesture-handler';
import React from 'react';
import {StatusBar} from 'react-native';
import {ApplicationProvider, IconRegistry} from '@ui-kitten/components';
import {EvaIconsPack} from '@ui-kitten/eva-icons';
import * as eva from '@eva-design/eva';
import {Provider as StoreProvider} from 'overmind-react';
import {store} from '@overmind/index';
import {Colors} from 'utils';
import RootNavigation from 'navigations/root-navigation';
import {default as theme} from './custom-theme.json'; // <-- Import app theme
import {AlertNotificationRoot} from 'react-native-alert-notification';
import * as Sentry from '@sentry/react-native';

/**
 * Use any valid `name` property from eva icons (e.g `github`, or `heart-outline`)
 * https://akveo.github.io/eva-icons
 */
Sentry.init({
  dsn: 'https://4621bb09fcc359ce9566eb6ab4c04d37@o4507185233461248.ingest.us.sentry.io/4507185235099648',
  // Set tracesSampleRate to 1.0 to capture 100% of transactions for performance monitoring.
  // We recommend adjusting this value in production.
  tracesSampleRate: 1.0,
  _experiments: {
    // profilesSampleRate is relative to tracesSampleRate.
    // Here, we'll capture profiles for 100% of transactions.
    profilesSampleRate: 1.0,
  },
});
function App(): React.ReactElement {
  return (
    <>
      <StoreProvider value={store}>
        <IconRegistry icons={EvaIconsPack} />
        <ApplicationProvider {...eva} theme={{...eva.light, ...theme}}>
          <AlertNotificationRoot>
            <StatusBar barStyle="dark-content" backgroundColor={Colors.WHITE} />
            <RootNavigation />
          </AlertNotificationRoot>
        </ApplicationProvider>
      </StoreProvider>
    </>
  );
}
export default Sentry.wrap(App);

// export default (): React.ReactElement => (
//   <>
//     <StoreProvider value={store}>
//       <IconRegistry icons={EvaIconsPack} />
//       <ApplicationProvider {...eva} theme={{...eva.light, ...theme}}>
//         <AlertNotificationRoot>
//           <StatusBar barStyle="dark-content" backgroundColor={Colors.WHITE} />
//           <RootNavigation />
//         </AlertNotificationRoot>
//       </ApplicationProvider>
//     </StoreProvider>
//   </>
// );
