/* eslint-disable react-native/no-inline-styles */
import {
  Image,
  ImageBackground,
  ScrollView,
  StatusBar,
  StyleSheet,
  Text,
  View,
} from 'react-native';
import React from 'react';
import {Colors, Mixins} from '@utils/index';
import images from '@assets/images';
import Login from './login';
import {IS_DEV} from '@env';

// import moment from 'moment';
const pkg = require('../../../package.json');

type Props = {
  navigation: any;
};

const Layout = (props: Props) => {
  const {navigation} = props;
  const now = new Date();

  return (
    <>
      <StatusBar barStyle="dark-content" backgroundColor={Colors.WHITE} />
      <View style={styles.container}>
        <ImageBackground
          source={images.BgLogin}
          resizeMode="cover"
          style={styles.image}>
          <View style={styles.overlay} />
          <View style={styles.content}>
            {/* <ScrollView> */}
            <View style={styles.header}>
              <Image style={styles.logo} source={images.logoLg} />
            </View>
            <View style={styles.formContainer}>
              <Image style={styles.busImage} source={images.Bus1Login} />
              <Image
                style={{
                  position: 'absolute',
                  width: Mixins.scaleSize(50),
                  height: Mixins.scaleSize(100),
                  bottom: Mixins.scaleSize(-50),
                  right: Mixins.scaleSize(0),
                }}
                source={images.ellipse1}
              />
              <ScrollView>
                <Login navigation={navigation} />
              </ScrollView>
            </View>
            {/* </ScrollView> */}
            <View style={styles.footer}>
              <Text style={styles.version}>
                Copyright {'\u00A9'} {now.getFullYear()} PT. Bagong, All Rights
                Reserved.
              </Text>
              <Text style={styles.version}>
                Version {pkg.version} {IS_DEV === '1' ? 'Development' : ''}
              </Text>
            </View>
          </View>
        </ImageBackground>
      </View>
    </>
  );
};

export default Layout;

const styles = StyleSheet.create({
  container: {
    flex: 1,
    backgroundColor: Colors.WHITE,
  },
  image: {
    flex: 1,
    // justifyContent: 'center',
  },
  overlay: {
    position: 'absolute',
    top: 0,
    right: 0,
    bottom: 0,
    left: 0,
    backgroundColor: '#ffffff',
    opacity: 0.3,
  },
  content: {
    flex: 1,
    position: 'relative',
    marginTop: Mixins.scaleSize(50),
  },
  header: {
    justifyContent: 'center',
    flexDirection: 'column',
    marginBottom: Mixins.scaleSize(50),
  },
  logo: {
    width: Mixins.WINDOW_WIDTH - 150,
    height: Mixins.WINDOW_WIDTH - 330,
    alignSelf: 'center',
    resizeMode: 'contain',
  },
  busImage: {
    width: Mixins.scaleSize(150),
    height: Mixins.scaleSize(150),
    position: 'absolute',
    right: 0,
    top: Mixins.scaleSize(-60),
  },
  formContainer: {
    flex: 1,
    height: '100%',
    borderTopLeftRadius: Mixins.scaleSize(30),
    borderTopRightRadius: Mixins.scaleSize(30),
    backgroundColor: Colors.WHITE,
    // paddingTop: Mixins.scaleSize(20),
    // paddingHorizontal: Mixins.scaleSize(20),
  },
  version: {
    color: Colors.WHITE,
    textAlign: 'center',
  },
  footer: {
    padding: Mixins.scaleSize(10),
    backgroundColor: '#40444B',
  },
});
