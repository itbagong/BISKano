declare module '*.svg' {
  import {SvgProps} from 'react-native-svg';
  const content: React.ComponentClass<SvgProps>;
  export default content;
}
